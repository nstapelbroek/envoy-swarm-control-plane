package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
)

const storageOperationTimeout = 3

type ObjectStorage struct {
	bucketName string
	cache      *DiskStorage
	client     *minio.Client
}

func NewObjectStorage(client *minio.Client, bucket string, cache *DiskStorage) *ObjectStorage {
	return &ObjectStorage{
		bucketName: bucket,
		client:     client,
		cache:      cache,
	}
}

func (o *ObjectStorage) GetStorageDirectory() string {
	return o.bucketName
}

func (o *ObjectStorage) GetFile(objectName string) (contents []byte, err error) {
	// read through cache
	if contents, err = o.cache.GetFile(objectName); err == nil {
		return contents, err
	}

	err = o.getAndCacheFile(objectName)
	if err != nil {
		return nil, err
	}

	return o.cache.GetFile(objectName)
}

func (o *ObjectStorage) PutFile(ObjectName string, contents []byte) (err error) {
	err = o.cache.PutFile(ObjectName, contents)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), storageOperationTimeout*time.Second)
	defer cancel()

	_, err = o.client.FPutObject(
		ctx,
		o.bucketName,
		ObjectName,
		fmt.Sprintf("%s/%s", o.cache.GetStorageDirectory(), ObjectName),
		minio.PutObjectOptions{},
	)

	return err
}

func (o *ObjectStorage) getAndCacheFile(fileName string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), storageOperationTimeout*time.Second)
	defer cancel()

	return o.client.FGetObject(
		ctx,
		o.bucketName,
		fileName,
		fmt.Sprintf("%s/%s", o.cache.GetStorageDirectory(), fileName), // needs proper calling of PutFile in the future
		minio.GetObjectOptions{},
	)
}
