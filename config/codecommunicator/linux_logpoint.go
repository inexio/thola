package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
)

type linuxLogpointCommunicator struct {
	codeCommunicator
}

// GetDiskComponentStorages returns the cpu load of ios devices.
func (c *linuxLogpointCommunicator) GetDiskComponentStorages(ctx context.Context) ([]device.DiskComponentStorage, error) {
	return c.parent.GetDiskComponentStorages(ctx)
}
