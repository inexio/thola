package mapping

import (
	"github.com/google/go-cmp/cmp"
	"github.com/inexio/thola/core/vfs"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestMappings(t *testing.T) {
	fileName := "ifType.yaml"
	f, err := vfs.FileSystem.Open(filepath.Join("mappings", fileName))
	if !assert.NoError(t, err, "failed to open mappings file from virtual file system") {
		return
	}
	b, err := ioutil.ReadAll(f)
	if !assert.NoError(t, err, "failed to read all bytes from vfs file") {
		return
	}

	mCompare := make(map[string]string)
	err = yaml.Unmarshal(b, mCompare)
	if !assert.NoError(t, err, "failed to unmarshal mapping file contents to map[string]string") {
		return
	}

	m, err := GetMapping(fileName)
	if assert.NoErrorf(t, err, "failed to get mapping of file %s", fileName) {
		assert.True(t, cmp.Equal(m, mCompare), "map returned by GetMapping() does not match the expected map")
	}

	_, err = GetMapping("file does not exist")
	assert.Error(t, err, "no error returned by GetMapping() for non existent mapping file")

	val, err := GetMappedValue(fileName, "1")
	if assert.NoErrorf(t, err, "failed to get mapping of file %s", fileName) {
		assert.Equal(t, "other", val, "wrong value returned by map")
	}

	_, err = GetMappedValue(fileName, "key does not exist")
	assert.Error(t, err, "no error returned by GetMappedValue() for non existent key")

	_, err = GetMappedValue("file does not exist", "key does not exist")
	assert.Error(t, err, "no error returned by GetMappedValue() for non existent mapping file")
}
