package tls

import listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

type Provider struct {
	tlsStorage     interface{}
	accountStorage interface{}
	email          string
}

func (p Provider) HasCertificate() bool {
	return false
}

func (p Provider) GetCertificate() interface{} {
	return false
}

func (p Provider) IssueCertificate() interface{} {
	return false
}

func (p Provider) UpgradeHttpListener(listener listener.Listener) interface{} {
	//	todo, implement this psuedo code:
	//	tlsListener := bootTlsListener()
	//	foreaxh vhost in http connectionmanager filter:
	//	if not tlsprovider.hascertificate(vhost.domains):
	//	tlsprovider.issueCertificate(vhost.domains)
	//	vhost = updateVhostWithIssueChallengeRoute()
	//	continue
	//}
	//
	//certificate = tlsprovicer.getCertificate(vhost.domains)
	//tlsListener.addVhost(vhost, certificate)
	//httpListener.replaceVhostWithTLSRedirect(vhost)
	return false
}
