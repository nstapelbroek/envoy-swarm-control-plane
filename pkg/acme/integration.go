package acme

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/go-acme/lego/v4/lego"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	tls "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
	"sync"
)

type Integration struct {
	http01Port    uint
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
		http01Port:    acmeChallengePort,
		http01Cluster: acmeClusterName,
		issueBacklog:  [][]string{},
		acmeEmail:     userEmail,
		certStorage:   certStorage,
		logger:        log,
	}
}

//PrepareVhostForIssuing will register and prepare the vhost for an ACME challenge
// note that the actual issuing is async
func (i *Integration) PrepareVhostForIssuing(vhost *route.VirtualHost) *route.VirtualHost {
	i.mutex.Lock()
	i.issueBacklog = append(i.issueBacklog, vhost.Domains)
	i.mutex.Unlock()

	vhost.Routes = append(vhost.Routes,
		&route.Route{
			Name: "acme_http01_route",
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: "/.well_known/",
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: i.http01Cluster,
					},
				},
			},
		})

	i.logger.WithFields(logger.Fields{"vhost": vhost.Name}).Debugf("Queued certificate issuing for vhost")

	return vhost
}

func (i *Integration) IssueCertificates() (reloadRequired bool, err error) {
	if len(i.issueBacklog) == 0 {
		return false, nil
	}

	return true, nil
}
