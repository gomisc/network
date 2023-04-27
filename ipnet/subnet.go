package ipnet

import (
	"net"
	"sync"
)

// IPFilterFunc - фильтр IP адресов
type IPFilterFunc func(addr net.IP) bool

// SubnetRange - имплементация контейнера IP адресов подсети
type SubnetRange struct {
	ipaddrs []string
	cidr    string

	mu   sync.Mutex
	used map[string]struct{}
}

// NewSubnetRage - возвращает контейнер адресов подсети
func NewSubnetRage(cidr string, filter IPFilterFunc) (*SubnetRange, error) {
	ip, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.Wrapf(err, "wrong CIDR %s", cidr)
	}

	var hosts []string

	for ip = ip.Mask(subnet.Mask); subnet.Contains(ip); ipIncrement(ip) {
		if filter(ip) {
			hosts = append(hosts, ip.String())
		}
	}

	return &SubnetRange{
		cidr:    cidr,
		ipaddrs: hosts,
		used:    make(map[string]struct{}),
	}, nil
}

// Range - возвращает список адресов подсети
func (r *SubnetRange) Range() []string {
	return r.ipaddrs
}

// Subnet - возвращает строковое значение подсети
func (r *SubnetRange) Subnet() string {
	return r.cidr
}

// NextIP - возвращает следующий не назначенный IP адрес подсети
func (r *SubnetRange) NextIP() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, ip := range r.ipaddrs {
		if _, ok := r.used[ip]; !ok {
			r.used[ip] = struct{}{}

			return ip
		}
	}

	return ""
}

func ipIncrement(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
