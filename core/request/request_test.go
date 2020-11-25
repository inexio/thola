package request

import (
	"github.com/inexio/thola/core/value"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckThresholds(t *testing.T) {
	th1 := CheckThresholds{
		WarningMin:  value.New(5),
		WarningMax:  value.New(10),
		CriticalMin: value.New(3),
		CriticalMax: value.New(12),
	}
	assert.NoError(t, th1.validate())

	th2 := CheckThresholds{}
	assert.NoError(t, th2.validate())

	th3 := CheckThresholds{
		WarningMax: value.New(3),
	}
	assert.NoError(t, th3.validate())

	th4 := CheckThresholds{
		WarningMin: value.New(2),
		WarningMax: value.New(1),
	}
	assert.Error(t, th4.validate())

	th5 := CheckThresholds{
		CriticalMin: value.New(2),
		CriticalMax: value.New(1),
	}
	assert.Error(t, th5.validate())

	th6 := CheckThresholds{
		WarningMin:  value.New(1),
		CriticalMin: value.New(2),
	}
	assert.Error(t, th6.validate())

	th7 := CheckThresholds{
		WarningMax:  value.New(2),
		CriticalMax: value.New(1),
	}
	assert.Error(t, th7.validate())
}
