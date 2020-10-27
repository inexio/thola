package tholaerr

import "github.com/pkg/errors"

type networkError interface {
	networkError() bool
}

// IsNetworkError returns true if an error is a network error.
func IsNetworkError(err error) bool {
	e, ok := errors.Cause(err).(networkError)
	return ok && e.networkError()
}

// SNMPError is an error returned by snmp functions.
type SNMPError struct {
	error
}

// NewSNMPError returns a new SNMPError
func NewSNMPError(msg string) error {
	return SNMPError{errors.New(msg)}
}

func (e SNMPError) networkError() bool {
	return true
}

// HTTPError is an error returned by snmp functions.
type HTTPError struct {
	error
}

// NewHTTPError returns an http error.
func NewHTTPError(msg string) error {
	return HTTPError{errors.New(msg)}
}

func (e HTTPError) networkError() bool {
	return true
}

type notFoundError interface {
	notFoundError() bool
}

// NotFoundError occurs when something is not found.
type NotFoundError struct {
	error
}

// NewNotFoundError returns an NotFoundError
func NewNotFoundError(msg string) error {
	return NotFoundError{errors.New(msg)}
}

func (e NotFoundError) notFoundError() bool {
	return true
}

// IsNotFoundError returns if the error is an NotFoundError
func IsNotFoundError(err error) bool {
	e, ok := errors.Cause(err).(notFoundError)
	return ok && e.notFoundError()
}

type preConditionError interface {
	preConditionError() bool
}

// PreConditionError occurs a condition for an action is not fulfilled.
type PreConditionError struct {
	error
}

// NewPreConditionError returns an PreConditionError
func NewPreConditionError(msg string) error {
	return PreConditionError{errors.New(msg)}
}

func (p PreConditionError) preConditionError() bool {
	return true
}

// IsPreConditionError returns if the error is an PreConditionError
func IsPreConditionError(err error) bool {
	e, ok := errors.Cause(err).(preConditionError)
	return ok && e.preConditionError()
}

type notImplementedError interface {
	notImplementedError() bool
}

// NotImplementedError occurs when a condition for an action is not fulfilled.
type NotImplementedError struct {
	error
}

// NewNotImplementedError returns an NotImplementedError
func NewNotImplementedError(msg string) error {
	return NotImplementedError{errors.New(msg)}
}

func (p NotImplementedError) notImplementedError() bool {
	return true
}

// IsNotImplementedError returns if the error is an NotImplementedError
func IsNotImplementedError(err error) bool {
	e, ok := errors.Cause(err).(notImplementedError)
	return ok && e.notImplementedError()
}

type tooManyRequestsError interface {
	tooManyRequestsError() bool
}

// TooManyRequestsError occurs when there were too many request sent to the api.
type TooManyRequestsError struct {
	error
}

// NewTooManyRequestsError returns an TooManyRequestsError
func NewTooManyRequestsError(msg string) error {
	return TooManyRequestsError{errors.New(msg)}
}

func (p TooManyRequestsError) tooManyRequestsError() bool {
	return true
}

// IsTooManyRequestsError returns if the error is an TooManyRequestsError
func IsTooManyRequestsError(err error) bool {
	e, ok := errors.Cause(err).(tooManyRequestsError)
	return ok && e.tooManyRequestsError()
}

type componentNotFound interface {
	componentNotFoundError() bool
}

// ComponentNotFoundError occurs when the specified component was not found
type ComponentNotFoundError struct {
	error
}

// NewComponentNotFoundError returns an ComponentNotFoundError
func NewComponentNotFoundError(msg string) error {
	return ComponentNotFoundError{errors.New(msg)}
}

func (p ComponentNotFoundError) componentNotFoundError() bool {
	return true
}

// IsComponentNotFound returns if the error is an ComponentNotFoundError
func IsComponentNotFound(err error) bool {
	e, ok := errors.Cause(err).(componentNotFound)
	return ok && e.componentNotFoundError()
}

// OutputError
//
// OutputError embeds all error messages which occur in requests on the API.
//
// swagger:model
type OutputError struct {
	Error string `json:"error" xml:"error"`
}
