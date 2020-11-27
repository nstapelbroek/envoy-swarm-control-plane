package tls

import (
	"context"
	"crypto/tls"
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
	configSource     *core.ConfigSource
	configKeyPrefix  string
	requestedConfigs map[string]*route.VirtualHost
	storage          *storage.Certificate
	logger           logger.Logger
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
		configSource:     c,
		configKeyPrefix:  "downstream_tls_",
		requestedConfigs: make(map[string]*route.VirtualHost),
		storage:          certificateStorage,
		logger:           log,
	}
}

func (p *CertificateSecretsProvider) HasValidCertificate(vhost *route.VirtualHost) bool {
	cert, err := p.getCertificate(vhost)
	if err != nil {
		return false
	}

	return IsCertUsable(cert)
}

// GetCertificateConfig will register vhost in the SDS mapping, assuring that the certificate is returned when calling Provide()
func (p *CertificateSecretsProvider) GetCertificateConfig(vhost *route.VirtualHost) *auth.SdsSecretConfig {
	key := p.getSecretConfigKey(vhost)
	p.requestedConfigs[key] = vhost

	return &auth.SdsSecretConfig{
		Name:      key,
		SdsConfig: p.configSource,
	}
}

func (p *CertificateSecretsProvider) Provide(_ context.Context) (secrets []types.Resource, err error) {
	for sdsKey := range p.requestedConfigs {
		vhost := p.requestedConfigs[sdsKey]

		// No need to re-validate anything at this point. We simply serve the bytes that are requested
		public, private, err := p.getCertificateFromStorage(vhost)
		if err != nil {
			p.logger.Warnf("promised certificate for %s is suddenly gone", sdsKey)
			continue
		}

		secrets = append(secrets, &auth.Secret{
			Name: sdsKey,
			Type: &auth.Secret_TlsCertificate{
				TlsCertificate: &auth.TlsCertificate{
					CertificateChain: &core.DataSource{
						// assuming that index 0 is the leaf
						Specifier: &core.DataSource_InlineBytes{InlineBytes: public},
					},
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: private},
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

// getCertificate retrieves a certificate from storage and parses it to assure it's usable
func (p *CertificateSecretsProvider) getCertificate(vhost *route.VirtualHost) (*tls.Certificate, error) {
	certBytes, keyBytes, err := p.getCertificateFromStorage(vhost)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		p.logger.Infof("decoding certificate from storage failed: %", err.Error())
		return nil, err
	}

	return &cert, err
}

func (p *CertificateSecretsProvider) getCertificateFromStorage(vhost *route.VirtualHost) ([]byte, []byte, error) {
	// First domain in the array is the primary one @see TestVhostPrimaryDomainIsFirstInDomains
	domains := vhost.GetDomains()
	if len(domains) == 0 {
		return nil, nil, errors.New("vhost contains no domains")
	}

	return p.storage.GetCertificate(vhost.GetDomains()[0], vhost.GetDomains())
}
