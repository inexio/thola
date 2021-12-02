package request

// CheckDeviceRequest
//
// CheckDeviceRequest is the request struct for the check device request.
//
// swagger:model
type CheckDeviceRequest struct {
	BaseRequest
	CheckRequest
}

func (r *CheckDeviceRequest) HandlePreProcessError(err error) (Response, error) {
	return r.CheckRequest.HandlePreProcessError(err)
}
