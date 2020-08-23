package s3

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type CertificateStorage struct {
	client *minio.Client
}

func (c *CertificateStorage) GetCertificate(domain string, sans []string) (publicChain, privateKey []byte, err error) {
	panic("implement me")
}

func NewCertificateStorage(storageEndpoint, storageAccessKey, storageSecretKey string) (*CertificateStorage, error) {
	minioClient, err := minio.New(storageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(storageAccessKey, storageSecretKey, ""),
		Secure: true,
	})

	if err != nil {
		return nil, err
	}

	return &CertificateStorage{
		client: minioClient,
	}, nil
}
