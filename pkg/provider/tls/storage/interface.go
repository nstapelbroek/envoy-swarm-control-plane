package storage

type CertificateStorage interface {
	GetCertificate(domain string, sans []string) (publicChain, privateKey []byte, err error)
}
