package swarm

import (
	"context"
	"testing"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/swarm/converting"
	"github.com/stretchr/testify/assert"
)

type hasAllSDS struct{}

func (k *hasAllSDS) Provide(_ context.Context) (secrets []types.Resource, err error) {
	panic("implement me")
}

func (k *hasAllSDS) HasValidCertificate(_ *route.VirtualHost) bool {
	return true
}

func (k *hasAllSDS) GetCertificateConfig(_ *route.VirtualHost) *auth.SdsSecretConfig {
	return &auth.SdsSecretConfig{}
}

func TestNewListenerBuilderAcceptsNilValues(t *testing.T) {
	// As this makes testing easier I opted for a test to guard us from changing this behaviour unintentionally
	result := NewListenerProvider(nil, nil)

	assert.IsType(t, &ListenerProvider{}, result)
}

func TestListenerBuilder_ProvideListeners(t *testing.T) {
	subject := NewListenerProvider(nil, nil)
	testcase := converting.NewVhostCollection()
	testcase.Vhosts["somedomain.com"] = &route.VirtualHost{
		Name:    "somedomain.com",
		Domains: []string{"somedomain.com", "www.somedomain.com"},
	}

	result, _ := subject.ProvideListeners(testcase)

	assert.Len(t, result, 1)
}

func TestListenerBuilder_createListenersFromVhostsEmptyVhostCollection(t *testing.T) {
	subject := NewListenerProvider(nil, nil)
	testcase := converting.NewVhostCollection()

	httpResult, httpsResult := subject.createListenersFromVhosts(testcase)

	assert.Len(t, httpResult.GetFilterChains(), 1)
	assert.Len(t, httpsResult.GetFilterChains(), 0)
}

func TestListenerBuilder_createListenersFromVhostsNoCerts(t *testing.T) {
	subject := NewListenerProvider(nil, nil)
	testcase := converting.NewVhostCollection()
	testcase.Vhosts["somedomain.com"] = &route.VirtualHost{
		Name:    "somedomain.com",
		Domains: []string{"somedomain.com", "www.somedomain.com"},
	}

	httpResult, httpsResult := subject.createListenersFromVhosts(testcase)

	assert.Len(t, httpResult.GetFilterChains(), 1)
	assert.Len(t, httpsResult.GetFilterChains(), 0)
}

func TestListenerBuilder_createListenersFromVhostsWithCerts(t *testing.T) {
	sds := hasAllSDS{}
	subject := NewListenerProvider(&sds, nil)
	testcase := converting.NewVhostCollection()
	testcase.Vhosts["somedomain.com"] = &route.VirtualHost{
		Name:    "somedomain.com",
		Domains: []string{"somedomain.com", "www.somedomain.com"},
	}

	httpResult, httpsResult := subject.createListenersFromVhosts(testcase)

	assert.Len(t, httpResult.GetFilterChains(), 1)
	assert.Len(t, httpsResult.GetFilterChains(), 1)
}

func TestListenerBuilder_createListenersFromMultiVhostsWithCerts(t *testing.T) {
	sds := hasAllSDS{}
	subject := NewListenerProvider(&sds, nil)
	testcase := converting.NewVhostCollection()
	testcase.Vhosts["somedomain.com"] = &route.VirtualHost{
		Name:    "somedomain.com",
		Domains: []string{"somedomain.com"},
	}
	testcase.Vhosts["anotherdomain.com"] = &route.VirtualHost{
		Name:    "anotherdomain.com",
		Domains: []string{"anotherdomain.com"},
	}
	testcase.Vhosts["example.nl"] = &route.VirtualHost{
		Name:    "example.nl",
		Domains: []string{"example.nl"},
	}

	httpResult, httpsResult := subject.createListenersFromVhosts(testcase)

	// clearly shows that every vhost with a certificate ends up as a separate filter chain for TLS connections
	assert.Len(t, httpResult.GetFilterChains(), 1)
	assert.Len(t, httpsResult.GetFilterChains(), 3)
}

func Test_createHTTPSRedirectVhost(t *testing.T) {
	originalVhost := &route.VirtualHost{
		Name:    "orignal.com",
		Domains: []string{"original.com", "www.original.com"},
	}

	redirectVhost := createHTTPSRedirectVhost(originalVhost)

	assert.Equal(t, originalVhost.Name, redirectVhost.Name)
	assert.Equal(t, originalVhost.Domains, redirectVhost.Domains)
	assert.NotEqual(t, originalVhost.Routes, redirectVhost.Routes)
	assert.Len(t, redirectVhost.Routes, 1)
	assert.Equal(t, "https_redirect", redirectVhost.Routes[0].Name)
	assert.Equal(t, &route.RouteMatch_Prefix{Prefix: "/"}, redirectVhost.Routes[0].Match.PathSpecifier)
	assert.IsType(t, &route.Route_Redirect{}, redirectVhost.Routes[0].Action)
}
