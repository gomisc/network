package ports

import (
	"fmt"
	"os"

	"github.com/onsi/ginkgo"
)

func (t TestPortRange) StartPortForNode() int {
	const startEphemeral, endEphemeral = 32768, 60999

	port := int(t) + portsPerNode*(ginkgo.GinkgoParallelNode()-1)
	if port >= startEphemeral-portsPerNode && port <= endEphemeral-portsPerNode {
		fmt.Fprintf(os.Stderr, "WARNING: port %d is part of the default ephemeral port range on linux", port)
	}

	return port
}
