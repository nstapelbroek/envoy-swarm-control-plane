package tls

import (
	"context"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
)

type CertificateSecretsProvider struct {
	configSource  *core.ConfigSource
	secretsPrefix string
	logger        logger.Logger
}

func NewCertificateSecretsProvider(controlPlaneClusterName string, log logger.Logger) *CertificateSecretsProvider {
	// we can re-use the config source for all secrets so we initialize it once :)
	c := &core.ConfigSource{
		ResourceApiVersion: core.ApiVersion_V3,

		// somehow this is not supported
		//ConfigSourceSpecifier: &core.ConfigSource_Self{
		//	Self: &core.SelfConfigSource{TransportApiVersion: core.ApiVersion_V3},
		//},
		ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &core.ApiConfigSource{
				ApiType:             core.ApiConfigSource_GRPC,
				TransportApiVersion: core.ApiVersion_V3,
				GrpcServices: []*core.GrpcService{{
					TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
							ClusterName: controlPlaneClusterName,
						},
					},
				}},
			},
		},
	}

	return &CertificateSecretsProvider{
		configSource:  c,
		secretsPrefix: "downstream_tls_",
		logger:        log,
	}
}

func (p *CertificateSecretsProvider) HasCertificate(vhost *route.VirtualHost) bool {
	return true
}

func (p *CertificateSecretsProvider) GetCertificateConfig(vhost *route.VirtualHost) *auth.SdsSecretConfig {
	key := p.getSecretConfigKey(vhost)

	return &auth.SdsSecretConfig{
		Name:      key,
		SdsConfig: p.configSource,
	}
}

func (p *CertificateSecretsProvider) Provide(ctx context.Context) (secrets []types.Resource, err error) {
	return []types.Resource{
		&auth.Secret{
			Name: snakeOilConfigKey,
			Type: &auth.Secret_TlsCertificate{
				TlsCertificate: &auth.TlsCertificate{
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(snakeOilPrivateKey)},
					},
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(snakeOilCertificate)},
					},
				},
			},
		},
	}, nil
}

func (p *CertificateSecretsProvider) getSecretConfigKey(vhost *route.VirtualHost) string {
	return snakeOilConfigKey
}
