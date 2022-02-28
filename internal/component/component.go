package component

import (
	"fmt"
	"github.com/pkg/errors"
)

// Component represents a component with a byte.
type Component byte

// All component enums.
const (
	Interfaces Component = iota + 1
	UPS
	CPU
	Memory
	SBC
	Server
	Disk
	HardwareHealth
	HighAvailability
	SIEM
)

// CreateComponent creates a component.
func CreateComponent(component string) (Component, error) {
	switch component {
	case "interfaces":
		return Interfaces, nil
	case "ups":
		return UPS, nil
	case "cpu":
		return CPU, nil
	case "memory":
		return Memory, nil
	case "sbc":
		return SBC, nil
	case "server":
		return Server, nil
	case "disk":
		return Disk, nil
	case "hardware_health":
		return HardwareHealth, nil
	case "high_availability":
		return HighAvailability, nil
	case "siem":
		return SIEM, nil
	default:
		return 0, fmt.Errorf("invalid component type: %s", component)
	}
}

// ToString returns the component as a string.
func (d *Component) ToString() (string, error) {
	if d == nil {
		return "", errors.New("component is empty")
	}
	switch *d {
	case Interfaces:
		return "interfaces", nil
	case UPS:
		return "ups", nil
	case CPU:
		return "cpu", nil
	case Memory:
		return "memory", nil
	case SBC:
		return "sbc", nil
	case Server:
		return "server", nil
	case Disk:
		return "disk", nil
	case HardwareHealth:
		return "hardware_health", nil
	case HighAvailability:
		return "high_availability", nil
	case SIEM:
		return "siem", nil
	default:
		return "", errors.New("unknown component")
	}
}
