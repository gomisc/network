package ipnet

import (
	"bufio"
	"encoding/binary"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	procFile  = "/proc/net/route"
	sep       = "\t"
	hexPrefix = "0x"
	zeroAddr  = "00000000"

	dstField  = 1
	gwField   = 2
	maskField = 7
)

// nolint: gosec
func getDefaultGateway() (net.IP, error) {
	file, err := os.Open(procFile)
	if err != nil {
		return nil, errors.Wrap(err, "read proc route file")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), sep)

		if tokens[dstField] == zeroAddr && tokens[maskField] == zeroAddr {
			return parseIP(tokens[gwField])
		}
	}

	return nil, ErrGatewayNotFound
}

func parseIP(str string) (net.IP, error) {
	dig, err := strconv.ParseInt(hexPrefix+str, 0, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parse address int")
	}

	ip := make(net.IP, net.IPv4len)
	binary.LittleEndian.PutUint32(ip, uint32(dig))

	return net.IPv4(ip[0], ip[1], ip[2], ip[3]), nil
}
