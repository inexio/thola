package request

import "github.com/inexio/thola/internal/device"

// ReadHighAvailabilityRequest
//
// ReadHighAvailabilityRequest is the request struct for the read high availability request.
//
// swagger:model
type ReadHighAvailabilityRequest struct {
	ReadRequest
}

// ReadHighAvailabilityResponse
//
// ReadHighAvailabilityResponse is the response struct for the read high availability request.
//
// swagger:model
type ReadHighAvailabilityResponse struct {
	HighAvailability device.HighAvailabilityComponent `yaml:"high_availability" json:"high_availability" xml:"high_availability"`
	ReadResponse
}
