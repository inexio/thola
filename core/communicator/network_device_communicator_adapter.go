package communicator

import "context"

type adapterFunc func(...interface{}) (interface{}, error)

type adapter struct {
	com availableCommunicatorFunctions
}

type communicatorAdapter interface {
	getVendor(...interface{}) (interface{}, error)
	getModel(...interface{}) (interface{}, error)
	getModelSeries(...interface{}) (interface{}, error)
	getSerialNumber(...interface{}) (interface{}, error)
	getOSVersion(...interface{}) (interface{}, error)
	getIfTable(...interface{}) (interface{}, error)
	getInterfaces(...interface{}) (interface{}, error)
	getCountInterfaces(...interface{}) (interface{}, error)
	communicatorAdapterUPS
}

type communicatorAdapterUPS interface {
	getUPSComponentAlarm(...interface{}) (interface{}, error)
	getUPSComponentAlarmLowVoltageDisconnect(...interface{}) (interface{}, error)
	getUPSComponentBatteryAmperage(...interface{}) (interface{}, error)
	getUPSComponentBatteryCapacity(...interface{}) (interface{}, error)
	getUPSComponentBatteryCurrent(...interface{}) (interface{}, error)
	getUPSComponentBatteryRemainingTime(...interface{}) (interface{}, error)
	getUPSComponentBatteryTemperature(...interface{}) (interface{}, error)
	getUPSComponentBatteryVoltage(...interface{}) (interface{}, error)
	getUPSComponentCurrentLoad(...interface{}) (interface{}, error)
	getUPSComponentMainsVoltageApplied(...interface{}) (interface{}, error)
	getUPSComponentRectifierCurrent(...interface{}) (interface{}, error)
	getUPSComponentSystemVoltage(...interface{}) (interface{}, error)
}

var emptyAdapterFunc adapterFunc

func newCommunicatorAdapter(com availableCommunicatorFunctions) communicatorAdapter {
	return &adapter{com}
}

func (a *adapter) getVendor(i ...interface{}) (interface{}, error) {
	return a.com.GetVendor(i[0].(context.Context))
}

func (a *adapter) getModel(i ...interface{}) (interface{}, error) {
	return a.com.GetModel(i[0].(context.Context))
}

func (a *adapter) getModelSeries(i ...interface{}) (interface{}, error) {
	return a.com.GetModelSeries(i[0].(context.Context))
}

func (a *adapter) getSerialNumber(i ...interface{}) (interface{}, error) {
	return a.com.GetSerialNumber(i[0].(context.Context))
}

func (a *adapter) getOSVersion(i ...interface{}) (interface{}, error) {
	return a.com.GetOSVersion(i[0].(context.Context))
}

func (a *adapter) getIfTable(i ...interface{}) (interface{}, error) {
	return a.com.GetIfTable(i[0].(context.Context))
}

func (a *adapter) getInterfaces(i ...interface{}) (interface{}, error) {
	return a.com.GetInterfaces(i[0].(context.Context))
}

func (a *adapter) getCountInterfaces(i ...interface{}) (interface{}, error) {
	return a.com.GetCountInterfaces(i[0].(context.Context))
}

func (a *adapter) getUPSComponentAlarm(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentAlarm(i[0].(context.Context))
}

func (a *adapter) getUPSComponentAlarmLowVoltageDisconnect(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentAlarmLowVoltageDisconnect(i[0].(context.Context))
}

func (a *adapter) getUPSComponentBatteryAmperage(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentBatteryAmperage(i[0].(context.Context))
}

func (a *adapter) getUPSComponentBatteryCapacity(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentBatteryCapacity(i[0].(context.Context))
}

func (a *adapter) getUPSComponentBatteryCurrent(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentBatteryCurrent(i[0].(context.Context))
}

func (a *adapter) getUPSComponentBatteryRemainingTime(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentBatteryRemainingTime(i[0].(context.Context))
}

func (a *adapter) getUPSComponentBatteryTemperature(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentBatteryTemperature(i[0].(context.Context))
}

func (a *adapter) getUPSComponentBatteryVoltage(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentBatteryVoltage(i[0].(context.Context))
}

func (a *adapter) getUPSComponentCurrentLoad(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentCurrentLoad(i[0].(context.Context))
}

func (a *adapter) getUPSComponentMainsVoltageApplied(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentMainsVoltageApplied(i[0].(context.Context))
}

func (a *adapter) getUPSComponentRectifierCurrent(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentRectifierCurrent(i[0].(context.Context))
}

func (a *adapter) getUPSComponentSystemVoltage(i ...interface{}) (interface{}, error) {
	return a.com.GetUPSComponentSystemVoltage(i[0].(context.Context))
}
