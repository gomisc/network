package ports

import (
	"os"
	"strconv"

	"gopkg.in/gomisc/errors.v1"
)

const (
	// ErrUnsupportedNetworkType не поддерживаемый тип сети
	ErrUnsupportedNetworkType = errors.Const("unsupported network type")
)

// Настройки типов портов
type (
	PortName  string
	HostPorts map[PortName]uint16
	DebugPort uint16

	// TestPortRange - диапазон портов для интеграционных тестов
	TestPortRange int
)

// Наименования портов
const (
	ListenPort    PortName = "ListenPort"
	WebPort       PortName = "WebPort"
	APIPort       PortName = "APIPort"
	ZookeeperPort PortName = "ZookeeperPort"
	CollectorPort PortName = "CollectorPort"
	ZipkinPort    PortName = "ZipkinPort"
	DebugPortName PortName = "DebugPort"
	HealthPort    PortName = "HealthCheckPort"
)

// Настройки диапазонов портов для тестов сервисов
const (
	BaseDebugPort = 5000
	basePort      = 20000
	portsPerNode  = 10
	portsPerSuite = 50 * portsPerNode
)

// Диапазоны портов для тестов по сервисам
const (
	ToolsBasePort TestPortRange = basePort + portsPerSuite*iota
)

// Порты отладки сервисов
const (
	DefaultDebug DebugPort = BaseDebugPort + iota
)

func (p DebugPort) Port() uint16 {
	if p.Enabled() {
		return uint16(p)
	}

	return 0
}

func (p DebugPort) String() string {
	if p.Enabled() {
		return strconv.FormatUint(uint64(p), 10)
	}

	return ""
}

func (p DebugPort) Command() string {
	switch p {
	default:
		return ""
	}
}

func (p DebugPort) Enabled() bool {
	const enabled = "true"

	switch p {
	case DefaultDebug:
		return os.Getenv("DEFAULT_DEBUG") == enabled
	default:
		return false
	}
}
