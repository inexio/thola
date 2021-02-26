package mapping

import (
	"github.com/inexio/thola/config"
	"github.com/inexio/thola/core/tholaerr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"sync"
)

type mapping map[string]string

var mappings struct {
	sync.Once
	sync.Mutex

	mappings map[string]mapping
}

func (m mapping) get(key string) (string, error) {
	if model, ok := m[key]; ok {
		return model, nil
	}
	return "", tholaerr.NewNotFoundError("oid is not in model mapping")
}

func readMapping(file string) (mapping, error) {
	f, err := config.FileSystem.Open(filepath.Join("mappings", file))
	if err != nil {
		return nil, errors.New("failed to open mappings file")
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.New("failed to read mappings file")
	}

	m := make(map[string]string)

	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return nil, errors.New("failed to unmarshal sysObjectID mappings file")
	}
	return m, nil
}

func initializeMapping(file string) error {
	mappings.Do(func() {
		mappings.mappings = make(map[string]mapping)
	})
	if mappings.mappings == nil {
		return errors.New("Mappings were not initialized")
	}

	mappings.Lock()
	defer mappings.Unlock()

	_, ok := mappings.mappings[file]
	if !ok {
		m, err := readMapping(file)
		if err != nil {
			return errors.Wrap(err, "failed to read mapping")
		}
		mappings.mappings[file] = m
	}

	if mappings.mappings[file] == nil {
		return errors.New("Mapping was not initialized")
	}
	return nil
}

// GetMappedValue returns the value which the key is associated with in the specified file.
func GetMappedValue(file, key string) (string, error) {
	err := initializeMapping(file)
	if err != nil {
		return "", err
	}

	return mappings.mappings[file].get(key)
}

// GetMapping returns the mapping of the specified file.
func GetMapping(file string) (map[string]string, error) {
	err := initializeMapping(file)
	if err != nil {
		return nil, err
	}

	return mappings.mappings[file], nil
}
