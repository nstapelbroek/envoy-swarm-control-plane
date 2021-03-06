package acme

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	tlsstorage "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type Integration struct {
	acmeClient      *lego.Client
	acmeClusterName string
	issueBacklog    map[string][]string
	renewalList     map[string][]string
	mutex           sync.Mutex
	certStorage     *tlsstorage.Certificate
	logger          logger.Logger
}

func NewIntegration(client *lego.Client, cluster string, certStorage *tlsstorage.Certificate, log logger.Logger) *Integration {
	return &Integration{
		acmeClient:      client,
		acmeClusterName: cluster,
		issueBacklog:    make(map[string][]string),
		renewalList:     make(map[string][]string),
		certStorage:     certStorage,
		logger:          log,
	}
}

// EnableAutoRenewal will administer the current domains of the vhost to a watchlist that gets checked every day
func (i *Integration) EnableAutoRenewal(vhost *route.VirtualHost) {
	go i.addToRenewalList(vhost.GetDomains())
}

// IsScheduledForIssuing will tell if a vhost is about to partake in an ACME challenge
func (i *Integration) IsScheduledForIssuing(vhost *route.VirtualHost) bool {
	_, isScheduled := i.issueBacklog[vhost.GetDomains()[0]]
	return isScheduled
}

// PrepareVhostForIssuing will add the vhost to the issue backlog and update the vhost config for any ACME challenge
func (i *Integration) PrepareVhostForIssuing(vhost *route.VirtualHost) *route.VirtualHost {
	vhost.Routes = append([]*route.Route{{
		Name: "acme_http01_route",
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: "/.well-known", // path prefix only works on first level at this moment
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: i.acmeClusterName,
				},
			},
		},
	}}, vhost.Routes...)

	// Prevent waiting for IssueCertificates() to complete
	go i.addToIssueBacklog(vhost.Domains)

	// See https://github.com/envoyproxy/envoy/issues/886, Host headers with a port value cause a mismatch
	// Unsure if this happens in the wild, but to be sure I'll update the vhost domains
	remappedDomains := make([]string, len(vhost.Domains)*2) //nolint:gomnd // it's twice the size of the original array
	copy(remappedDomains, vhost.Domains)
	for i := range vhost.Domains {
		remappedDomains[len(vhost.Domains)+i] = fmt.Sprintf("%s:80", vhost.Domains[i])
	}

	vhost.Domains = remappedDomains
	i.logger.WithFields(logger.Fields{"vhost": vhost.Name}).Debugf("vhost configured for ACME issuing")

	return vhost
}

func (i *Integration) IssueCertificates() (reloadRequired bool, err error) {
	if len(i.issueBacklog) == 0 {
		i.logger.Debugf("No certificates to issue")
		return false, nil
	}

	// Prevent edge cases by locking our data
	i.mutex.Lock()
	for primaryDomain := range i.issueBacklog {
		domains := i.issueBacklog[primaryDomain]

		request := certificate.ObtainRequest{Domains: domains, Bundle: true}
		certs, err := i.acmeClient.Certificate.Obtain(request)
		if err != nil {
			i.logger.Errorf("failed issuing certificate: %s", err.Error())
			delete(i.issueBacklog, primaryDomain)
			continue
		}

		if err = i.certStorage.PutCertificate(domains[0], domains, certs.Certificate, certs.PrivateKey); err != nil {
			i.logger.Errorf("failed saving certificate to storage: %s", err.Error())
			// I don't know how to recover from this? Assuming I can re-obtain the cert another round
			delete(i.issueBacklog, primaryDomain)
			continue
		}

		delete(i.issueBacklog, primaryDomain)
		reloadRequired = true
	}
	i.mutex.Unlock()

	return reloadRequired, err
}

func (i *Integration) ScheduleRenewals() (reloadRequired bool) {
	const CertificateExpiryThreshold = 720 // 720 hours = 30 days
	if len(i.renewalList) == 0 {
		i.logger.Debugf("No certificates to watch for renewal")
		return reloadRequired
	}

	for primaryDomain := range i.renewalList {
		domains := i.renewalList[primaryDomain]

		certBytes, keyBytes, err := i.certStorage.GetCertificate(primaryDomain, domains)
		if err != nil {
			i.logger.Warnf("skipped renewal check for %s due to storage error", primaryDomain)
			continue
		}

		pair, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			i.logger.Warnf("parsing certificate from storage failed: %", err.Error())
			continue
		}

		cert, _ := x509.ParseCertificate(pair.Certificate[0])
		if time.Now().Add(CertificateExpiryThreshold * time.Hour).After(cert.NotAfter) {
			go i.addToIssueBacklog(domains)
			i.logger.Infof("queued renewal of certificate for %s", primaryDomain)
			reloadRequired = true
		}
	}

	return reloadRequired
}

func (i *Integration) addToRenewalList(domains []string) {
	backlogKey := domains[0] // @see TestVhostPrimaryDomainIsFirstInDomains
	if _, exists := i.renewalList[backlogKey]; exists {
		return
	}

	i.mutex.Lock()
	i.renewalList[backlogKey] = domains
	i.mutex.Unlock()
}

func (i *Integration) addToIssueBacklog(domains []string) {
	backlogKey := domains[0] // @see TestVhostPrimaryDomainIsFirstInDomains
	if _, exists := i.issueBacklog[backlogKey]; exists {
		return
	}

	i.mutex.Lock()
	i.issueBacklog[backlogKey] = domains
	i.mutex.Unlock()
}
