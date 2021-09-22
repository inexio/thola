package hierarchy

import (
	"github.com/inexio/thola/internal/communicator"
)

// Hierarchy defines the hierarchy between multiple network device communicators.
// It mirrors the structure defined in the config/device-classes directory.
type Hierarchy struct {
	NetworkDeviceCommunicator communicator.Communicator
	Children                  map[string]Hierarchy
	TryToMatchLast            bool
}
