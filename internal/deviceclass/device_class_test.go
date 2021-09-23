package deviceclass

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeviceClass_GetHierarchy(t *testing.T) {
	_, err := GetHierarchy()
	assert.NoError(t, err, "hierarchy building failed")
}
