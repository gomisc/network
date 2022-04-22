//go:build !darwin && !linux
// +build !darwin,!linux

package ipnet

func getDefaultGateway() (net.IP, error) {
	return nil, ErrOSNotImplemented
}
