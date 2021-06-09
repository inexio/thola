package request

import "github.com/inexio/thola/internal/device"

// ReadDiskRequest
//
// ReadDiskRequest is a the request struct for the read disk request.
//
// swagger:model
type ReadDiskRequest struct {
	ReadRequest
}

// ReadDiskResponse
//
// ReadDiskResponse is a the response struct for the read disk response.
//
// swagger:model
type ReadDiskResponse struct {
	Disk device.DiskComponent `yaml:"disk" json:"disk" xml:"disk"`
	ReadResponse
}
