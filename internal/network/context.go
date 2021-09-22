package network

import "context"

type ctxKey byte

const (
	requestDeviceConnectionKey ctxKey = iota + 1
	snmpGetsInsteadOfWalk
)

// NewContextWithDeviceConnection returns a new context with the device connection
func NewContextWithDeviceConnection(ctx context.Context, con *RequestDeviceConnection) context.Context {
	return context.WithValue(ctx, requestDeviceConnectionKey, con)
}

// DeviceConnectionFromContext gets the device connection from the context
func DeviceConnectionFromContext(ctx context.Context) (*RequestDeviceConnection, bool) {
	con, ok := ctx.Value(requestDeviceConnectionKey).(*RequestDeviceConnection)
	return con, ok
}

// NewContextWithSNMPGetsInsteadOfWalk returns a new context with the request
func NewContextWithSNMPGetsInsteadOfWalk(ctx context.Context, b bool) context.Context {
	return context.WithValue(ctx, snmpGetsInsteadOfWalk, b)
}

// SNMPGetsInsteadOfWalkFromContext gets the request from the context
func SNMPGetsInsteadOfWalkFromContext(ctx context.Context) (bool, bool) {
	con, ok := ctx.Value(snmpGetsInsteadOfWalk).(bool)
	return con, ok
}
