package storage

type StorageMap struct {
	m map[string]string
}

func NewStorageMap() *StorageMap {
	return &StorageMap{
		m: make(map[string]string),
	}
}

func (s *StorageMap) GetURL(key string) string {
	return s.m[key]
}

func (s *StorageMap) PostURL(key string, ourl string) {
	s.m[key] = ourl
}
