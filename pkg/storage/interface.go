package storage

type Storage interface {
	GetStorageDirectory() string
	GetFile(fileName string) ([]byte, error)
	PutFile(fileName string, contents []byte) error
}
