package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
)

const fileMode = 0600

type DiskStorage struct {
	directory string
	logger    logger.Logger
}

func NewDiskStorage(path string, log logger.Logger) *DiskStorage {
	path = strings.TrimSuffix(path, "/")

	return &DiskStorage{directory: path, logger: log}
}

func (c *DiskStorage) GetFile(fileName string) (content []byte, err error) {
	content, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", c.directory, fileName))
	if err != nil {
		return nil, err
	}

	return content, err
}

func (c *DiskStorage) PutFile(fileName string, contents []byte) (err error) {
	log := c.getLogger(fileName)
	err = ioutil.WriteFile(
		fmt.Sprintf("%s/%s", c.directory, fileName),
		contents,
		os.FileMode(fileMode),
	)
	if err != nil {
		log.Warnf("error while writing file", err.Error())
	}

	return err
}

func (c *DiskStorage) getLogger(fileName string) logger.Logger {
	return c.logger.WithFields(logger.Fields{"driver": "disk", "fileName": fileName, "directory": c.directory})
}

func (c *DiskStorage) GetStorageDirectory() string {
	return c.directory
}
