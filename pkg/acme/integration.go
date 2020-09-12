package acme

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/go-acme/lego/v4/lego"
	acme "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/acme/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	tls "github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type Integration struct {
	http01Port  uint
	http01Route *route.Route
	lego        *lego.Client
	userStorage acme.User
	certStorage tls.Certificate
	logger      logger.Logger
}

func NewIntegration(port uint, userEmail, acmeClusterName string, certStorage tls.Certificate, log logger.Logger) *Integration {
	r := &route.Route{
		Name: "acme_http01_route",
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: "/.well_known/",
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: acmeClusterName,
				},
			},
		},
	}

	// We store the acme user keys in the same directory as the certificates for now
	u := acme.User{Storage: certStorage.Storage}

	return &Integration{http01Port: port, http01Route: r, userStorage: u, certStorage: certStorage, logger: log}
}

func (i *Integration) GetHTTP01Route() *route.Route {
	return i.http01Route
}
