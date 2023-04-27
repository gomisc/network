package ipnet

import (
	"context"
	"net"
	"sync"
)

type (
	// NetworksSet - ключ IP сети, значение - битовая маска
	NetworksSet map[string]int
	// NetworksGetter - возвращает сет сетей
	NetworksGetter func(ctx context.Context) (NetworksSet, error)
)

// NetworksAllocator интерфейс резервирования сетей
type NetworksAllocator struct {
	addr   *net.IPNet
	getter NetworksGetter

	mu   sync.RWMutex
	used NetworksSet
}

// NewNetworkAllocator конструктор интерфейса резервирования сетей
func NewNetworkAllocator(getter NetworksGetter, reserved ...string) (*NetworksAllocator, error) {
	reservedNetworks, err := getReservedNetworks(reserved...)
	if err != nil {
		return nil, errors.Wrap(err, "fill reserved networks")
	}

	na := &NetworksAllocator{
		getter: getter,
		used:   reservedNetworks,
		addr: &net.IPNet{
			IP:   net.IPv4(172, 16, 0, 0),
			Mask: net.IPv4Mask(255, 240, 0, 0),
		},
	}

	return na, nil
}

// GetFreeSubnet возвращает адрес свободной сети
func (na *NetworksAllocator) GetFreeSubnet(ctx context.Context) (*net.IPNet, error) {
	na.mu.Lock()
	defer na.mu.Unlock()

	used, err := na.getter(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get used networks")
	}

	for addr, mask := range used {
		na.used[addr] = mask
	}

	max, _ := na.addr.Mask.Size()

	for ni := int(na.addr.IP.To4()[1]); ni < int(na.addr.IP.To4()[1])+max; ni++ {
		global := net.IPNet{
			IP:   net.IPv4(na.addr.IP.To4()[0], byte(ni), 0, 0),
			Mask: net.IPv4Mask(255, 255, 0, 0),
		}
		if sz, ok := na.used[global.String()]; !ok || sz != 16 {
			for si := 0; si < 255; si++ {
				subnet := net.IPNet{
					IP:   net.IPv4(na.addr.IP.To4()[0], global.IP.To4()[1], byte(si), 0),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				}
				if _, exist := na.used[subnet.String()]; !exist {
					na.used[subnet.String()] = 24
					return &subnet, nil
				}
			}
		}
	}
	return nil, nil
}

func getReservedNetworks(reserved ...string) (NetworksSet, error) {
	set := make(NetworksSet)

	for i := 0; i < len(reserved); i++ {
		_, reserve, err := net.ParseCIDR(reserved[i])
		if err != nil {
			return nil, errors.Ctx().Str("parsed", reserved[i]).Wrap(err, "parse reserved network")
		}

		sz, _ := reserve.Mask.Size()
		set[reserve.String()] = sz
	}

	return set, nil
}
