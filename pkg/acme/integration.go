package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"strconv"
	"sync"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	tls "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type Integration struct {
	http01Port    string
	http01Cluster string
	acmeEmail     string
	issueBacklog  [][]string
	mutex         sync.Mutex
	lego          *lego.Client
	certStorage   tls.Certificate
	logger        logger.Logger
}

func NewIntegration(userEmail, acmeClusterName string, acmeChallengePort uint, certStorage tls.Certificate, log logger.Logger) *Integration {
	return &Integration{
		http01Port:    strconv.Itoa(int(acmeChallengePort)),
		http01Cluster: acmeClusterName,
		issueBacklog:  [][]string{},
		acmeEmail:     userEmail,
		certStorage:   certStorage,
		logger:        log,
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
					Cluster: i.http01Cluster,
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

type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (i *Integration) IssueCertificates() (reloadRequired bool, err error) {
	if len(i.issueBacklog) == 0 {
		return false, nil
	}

	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		i.logger.Fatalf(err.Error())
	}

	myUser := MyUser{
		Email: "you@yours.com",
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = "https://localhost:14000/dir"
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		i.logger.Fatalf(err.Error())
	}

	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("0.0.0.0", i.http01Port))
	if err != nil {
		i.logger.Fatalf(err.Error())
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		i.logger.Fatalf(err.Error())
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{"example.com"},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		i.logger.Fatalf(err.Error())
	}

	if certificates == nil || certificates.PrivateKey == nil || certificates.Certificate == nil {
		i.logger.Fatalf("well darn, no certificates")
	}

	i.certStorage.PutCertificate("mydomain.com", []string{}, certificates.Certificate, certificates.PrivateKey)

	return true, nil
}
