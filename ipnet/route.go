package ipnet

import (
	"net"
	"runtime"
)

const (
	ErrParseIP          = errors.Const("parse IP error")
	ErrGatewayNotFound  = errors.Const("can't get default gateway")
	ErrOSNotImplemented = errors.Const("not implemented for " + runtime.GOOS)
)

// GetDefaultGateway - возвращает де
func GetDefaultGateway() (net.IP, error) {
	return getDefaultGateway()
}
