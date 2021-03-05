package request

import (
	"github.com/inexio/thola/core/device"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckThresholds(t *testing.T) {
	th1 := CheckThresholds{
		WarningMin:  getPointer(5),
		WarningMax:  getPointer(10),
		CriticalMin: getPointer(3),
		CriticalMax: getPointer(12),
	}
	assert.NoError(t, th1.validate())

	th2 := CheckThresholds{}
	assert.NoError(t, th2.validate())

	th3 := CheckThresholds{
		WarningMax: getPointer(3),
	}
	assert.NoError(t, th3.validate())

	th4 := CheckThresholds{
		WarningMin: getPointer(2),
		WarningMax: getPointer(1),
	}
	assert.Error(t, th4.validate())

	th5 := CheckThresholds{
		CriticalMin: getPointer(2),
		CriticalMax: getPointer(1),
	}
	assert.Error(t, th5.validate())

	th6 := CheckThresholds{
		WarningMin:  getPointer(1),
		CriticalMin: getPointer(2),
	}
	assert.Error(t, th6.validate())

	th7 := CheckThresholds{
		WarningMax:  getPointer(2),
		CriticalMax: getPointer(1),
	}
	assert.Error(t, th7.validate())
}

func getPointer(f float64) *float64 {
	return &f
}

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
		interfaces = removeInterface(interfaces, toRemove, 0)
		assert.Equal(t, i/2, len(interfaces), "expected removed length and actual length differs")
	}
}
