package request

import (
	"github.com/inexio/thola/core/device"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveInterfaces(t *testing.T) {
	for i := 1; i < 1000; i++ {
		var interfaces []device.Interface
		var toRemove []int
		for j := 0; j < i; j++ {
			interfaces = append(interfaces, device.Interface{})
			if j%2 == 0 {
				toRemove = append(toRemove, j)
			}
		}
		interfaces = filterInterfaces(interfaces, toRemove, 0)
		assert.Equal(t, i/2, len(interfaces), "expected removed length and actual length differs")
	}
}
