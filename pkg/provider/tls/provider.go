package tls

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
