package tls

import (
	"context"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
)

type LetsEncryptProvider struct {
	registrationEmail string
	logger            logger.Logger
}

func NewLetsEncryptProvider(registrationEmail string, log logger.Logger) *LetsEncryptProvider {
	var email string
	if len(registrationEmail) != 0 {
		email = registrationEmail
	}

	return &LetsEncryptProvider{registrationEmail: email, logger: log}
}

func (p *LetsEncryptProvider) Provide(ctx context.Context) (secrets []types.Resource, err error) {
	return []types.Resource{
		&auth.Secret{
			Name: snakeOilConfigKey,
			Type: &auth.Secret_TlsCertificate{
				TlsCertificate: &auth.TlsCertificate{
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(privateKey)},
					},
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(privateChain)},
					},
				},
			},
		},
		&auth.Secret{
			Name: "neen",
			Type: &auth.Secret_ValidationContext{
				ValidationContext: &auth.CertificateValidationContext{
					TrustedCa: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{InlineBytes: []byte(rootCert)},
					},
				},
			},
		},
	}, nil
}

func (p *LetsEncryptProvider) GetCertificateConfigKey(vhost *route.VirtualHost) string {
	return snakeOilConfigKey
}
