package morningpost

import (
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path"

	"golang.org/x/exp/maps"
)

type FileStore struct {
	Data map[string]Feed
	Path string
}

func (f *FileStore) Add(feed Feed) {
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(feed.Endpoint)))
	feed.ID = id
	f.Data[id] = feed
}

func (f *FileStore) GetAll() []Feed {
	return maps.Values(f.Data)
}

func (f *FileStore) Load() error {
	file, err := os.Open(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(file)
	return dec.Decode(&f.Data)
}

func (f *FileStore) Save() error {
	file, err := os.Create(f.Path)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(file)
	return enc.Encode(f.Data)
}

func (f *FileStore) Delete(endpoint string) {
	delete(f.Data, endpoint)
}

func NewFileStore(opts ...FileStoreOption) (*FileStore, error) {
	fileStore := &FileStore{
		Data: map[string]Feed{},
		Path: userStateDir() + "/MorningPost/morningpost.db",
	}
	for _, o := range opts {
		o(fileStore)
	}
	err := fileStore.Load()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path.Dir(fileStore.Path)); os.IsNotExist(err) {
		err := os.MkdirAll(path.Dir(fileStore.Path), 0755)
		if err != nil {
			return nil, err
		}

	}
	return fileStore, nil
}

type FileStoreOption func(*FileStore)

func WithFileStorePath(path string) FileStoreOption {
	return func(f *FileStore) {
		f.Path = path
	}
}
