package tls

import (
	"context"
	"errors"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
)

type CertificateSecretsProvider struct {
	configSource          *core.ConfigSource
	configKeyPrefix       string
	requestedCertificates map[string]*route.VirtualHost
	storage               *storage.Certificate
	logger                logger.Logger
}

func NewCertificateSecretsProvider(controlPlaneClusterName string, certificateStorage *storage.Certificate, log logger.Logger) *CertificateSecretsProvider {
	// we can re-use the config source for all secrets so we initialize it once :)
	c := &core.ConfigSource{
		ResourceApiVersion: core.ApiVersion_V3,

		// somehow this is not supported
		// ConfigSourceSpecifier: &core.ConfigSource_Self{
		// 	 Self: &core.SelfConfigSource{TransportApiVersion: core.ApiVersion_V3},
		// },
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
		configSource:          c,
		configKeyPrefix:       "downstream_tls_",
		requestedCertificates: make(map[string]*route.VirtualHost),
		storage:               certificateStorage,
		logger:                log,
	}
}

func (p *CertificateSecretsProvider) HasCertificate(vhost *route.VirtualHost) bool {
	_, _, err := p.getCertificateFromStorage(vhost)
	return err == nil
}

// GetCertificateConfig will register vhost in the SDS mapping, assuring that the secrets will be available
func (p *CertificateSecretsProvider) GetCertificateConfig(vhost *route.VirtualHost) *auth.SdsSecretConfig {
	key := p.getSecretConfigKey(vhost)
	p.requestedCertificates[key] = vhost

	return &auth.SdsSecretConfig{
		Name:      key,
		SdsConfig: p.configSource,
	}
}

func (p *CertificateSecretsProvider) Provide(ctx context.Context) (secrets []types.Resource, err error) {
	for sdsKey := range p.requestedCertificates {
		vhost := p.requestedCertificates[sdsKey]

		// Assume that certificates are just there, no snake-oil fallback at this moment
		publicChain, privateKey, err := p.getCertificateFromStorage(vhost)
		if err != nil {
			continue
		}

		secrets = append(secrets, &auth.Secret{
			Name: sdsKey,
			Type: &auth.Secret_TlsCertificate{
				TlsCertificate: &auth.TlsCertificate{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: publicChain},
					},
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: privateKey},
					},
				},
			},
		})
	}

	return secrets, nil
}

func (p *CertificateSecretsProvider) getSecretConfigKey(vhost *route.VirtualHost) string {
	return p.configKeyPrefix + strings.ToLower(vhost.Name)
}

func (p *CertificateSecretsProvider) getCertificateFromStorage(vhost *route.VirtualHost) ([]byte, []byte, error) {
	// We will use the first domain in the slice as the primary domain name
	// conveniently this happens to to match the way we parse labels @see TestVhostPrimaryDomainIsFirstInDomains
	domains := vhost.GetDomains()
	if len(domains) == 0 {
		return nil, nil, errors.New("vhost contains no domains")
	}

	return p.storage.GetCertificate(vhost.GetDomains()[0], vhost.GetDomains())
}
