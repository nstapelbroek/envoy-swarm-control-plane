package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/provider/tls/storage/disk"
)

const storageOperationTimeout = 3

type CertificateStorage struct {
	bucketName string
	client     *minio.Client
	cache      *disk.CertificateStorage
}

func (c CertificateStorage) PutCertificate(domain string, sans []string, publicChain, privateKey []byte) error {
	panic("implement me")
}

func (c CertificateStorage) GetCertificate(domain string, sans []string) (publicChain, privateKey []byte, err error) {
	fileName := storage.GetCertificateFilename(domain, sans)

	// read through cache
	if publicChain, privateKey, err = c.cache.GetCertificate(domain, sans); err == nil {
		return publicChain, privateKey, err
	}

	// no object in cache, update if we can
	err = c.getAndCacheFile(fmt.Sprintf("%s.%s", fileName, storage.CertificateExtension))
	if err != nil {
		return nil, nil, err
	}

	err = c.getAndCacheFile(fmt.Sprintf("%s.%s", fileName, storage.PrivateKeyExtension))
	if err != nil {
		return nil, nil, err
	}

	return c.cache.GetCertificate(domain, sans)
}

func (c *CertificateStorage) getAndCacheFile(fileName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), storageOperationTimeout*time.Second)
	defer cancel()

	return c.client.FGetObject(
		ctx,
		c.bucketName,
		fileName,
		fmt.Sprintf("%s/%s", c.cache.GetStorageDirectory(), fileName),
		minio.GetObjectOptions{},
	)
}

func NewCertificateStorage(endpoint, bucket, access, key string, cache *disk.CertificateStorage) (*CertificateStorage, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, key, ""),
		Secure: true,
	})

	if err != nil {
		return nil, err
	}

	return &CertificateStorage{
		bucketName: bucket,
		client:     minioClient,
		cache:      cache,
	}, nil
}
