package storage

type CertificateStorage interface {
	GetCertificate(domains []string) (publicChain, privateKey []byte, err error)
}
