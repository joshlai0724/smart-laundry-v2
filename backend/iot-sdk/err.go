package iotsdk

import "errors"

var ErrRPCRequestTimeout = errors.New("rpc request timeout")

type InvalidParameterError struct {
	S string
}

func (e *InvalidParameterError) Error() string {
	return e.S
}

type DeviceNotFoundError struct {
	S string
}

func (e *DeviceNotFoundError) Error() string {
	return e.S
}

type StoreNotFoundError struct {
	S string
}

func (e *StoreNotFoundError) Error() string {
	return e.S
}

type InternalError struct {
	S string
}

func (e *InternalError) Error() string {
	return e.S
}

type IllegalStateError struct {
	S string
}

func (e *IllegalStateError) Error() string {
	return e.S
}
