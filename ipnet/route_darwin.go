package ipnet

import (
	"net"

	"golang.org/x/net/route"

	"git.eth4.dev/golibs/errors"
)

var defaultRoute = [4]byte{0, 0, 0, 0} // nolint

func getDefaultGateway() (net.IP, error) {
	rib, err := route.FetchRIB(0, route.RIBTypeRoute, 0)
	if err != nil {
		return nil, errors.Wrap(err, "fetch route information base")
	}

	var rms []route.Message

	rms, err = route.ParseRIB(route.RIBTypeRoute, rib)
	if err != nil {
		return nil, errors.Wrap(err, "parse route messages")
	}

	for i := 0; i < len(rms); i++ {
		if rm, isRouteMsg := rms[i].(*route.RouteMessage); isRouteMsg {
			dst, dstIsIPv4 := rm.Addrs[0].(*route.Inet4Addr)
			gw, gwIsIPv4 := rm.Addrs[1].(*route.Inet4Addr)

			if dstIsIPv4 && gwIsIPv4 && dst.IP == defaultRoute {
				return net.IPv4(gw.IP[0], gw.IP[1], gw.IP[2], gw.IP[3]), nil
			}
		}
	}

	return nil, ErrGatewayNotFound
}
