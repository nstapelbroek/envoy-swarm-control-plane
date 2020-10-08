package acme

import (
	"fmt"
	"sync"

	"github.com/go-acme/lego/v4/lego"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	tlsstorage "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type Integration struct {
	acmeClient      *lego.Client
	acmeClusterName string
	issueBacklog    [][]string
	mutex           sync.Mutex
	certStorage     *tlsstorage.Certificate
	logger          logger.Logger
}

func NewIntegration(client *lego.Client, cluster string, certStorage *tlsstorage.Certificate, log logger.Logger) *Integration {
	return &Integration{
		acmeClient:      client,
		acmeClusterName: cluster,
		issueBacklog:    [][]string{},
		certStorage:     certStorage,
		logger:          log,
	}
}

// PrepareVhostForIssuing will register and prepare the vhost for an ACME challenge
// note that the actual issuing is async
func (i *Integration) PrepareVhostForIssuing(vhost *route.VirtualHost) *route.VirtualHost {
	i.mutex.Lock()
	i.issueBacklog = append(i.issueBacklog, vhost.Domains)
	i.mutex.Unlock()

	// Prepend .well-known matcher
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

	// See https://github.com/envoyproxy/envoy/issues/886, clients using a port in their host header causes a mismatch
	// Unsure if this happens in the wild, but for local ACME testing I'll add the domains with port to
	// help requests find their way to the challenge server
	remappedDomains := make([]string, len(vhost.Domains)*2)
	copy(remappedDomains, vhost.Domains)
	for i := range vhost.Domains {
		remappedDomains[len(vhost.Domains)+i] = fmt.Sprintf("%s:80", vhost.Domains[i])
	}
	vhost.Domains = remappedDomains

	i.logger.WithFields(logger.Fields{"vhost": vhost.Name}).Debugf("Queued certificate issuing for vhost")

	return vhost
}

func (i *Integration) IssueCertificates() (reloadRequired bool, err error) {
	if len(i.issueBacklog) == 0 {
		return false, nil
	}

	// todo this is where i left off
	//for index := range i.issueBacklog {
	//	domains := i.issueBacklog[index]
	//	publicChain, privateKey, err := i.acmeClient.issueCertificate(i.issueBacklog[i])
	//	if err != nil {
	//		i.logger.Errorf(err.Error())
	//		continue
	//	}
	//
	//	_ = i.certStorage.PutCertificate(domains[0], domains, publicChain, privateKey)
	//	reloadRequired = true
	//}

	return reloadRequired, err
}
