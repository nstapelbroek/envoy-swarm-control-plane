package storage

type CertificateStorage interface {
	GetCertificate(domain string, sans []string) (publicChain, privateKey []byte, err error)
	PutCertificate(domain string, sans []string, publicChain, privateKey []byte) error
}
