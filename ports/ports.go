package ports

import (
	"os"
	"strconv"

	"git.corout.in/golibs/errors"
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
	TokenManagerBasePort
	MetricsGateBasePort
	MediationBasePort
	MetricsEnricherBasePort
	ProductCatalogBasePort
	ChargingBasePort
	ReportBasePort
	ForisPluginBasePort
	MonetizationBasePort
	ElasticPluginBasePort
	CDNPluginBasePort
)

// Порты отладки сервисов
const (
	APIGateDebug DebugPort = BaseDebugPort + iota
	MetricsGateDebug
	TokenManagerDebug
	MediationDebug
	MetricsEnricherDebug
	ProductCatalogDebug
	ChargingDebug
	ReportDebug
	ForisPluginDebug
	MonetizationDebug
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
	case APIGateDebug:
		return "/bin/api-gate"
	case MetricsGateDebug:
		return "/bin/metricsgate"
	case TokenManagerDebug:
		return "/bin/tokenmanager"
	case MediationDebug:
		return "/bin/mediation"
	case MetricsEnricherDebug:
		return "/bin/metricsenricher"
	case ProductCatalogDebug:
		return "/bin/productcatalog"
	case ChargingDebug:
		return "/bin/charging"
	case ReportDebug:
		return "/bin/report"
	case ForisPluginDebug:
		return "/bin/forisplugin"
	case MonetizationDebug:
		return "/bin/monetization"
	default:
		return ""
	}
}

func (p DebugPort) Enabled() bool {
	const enabled = "true"

	switch p {
	case APIGateDebug:
		return os.Getenv("API_GATE_DEBUG") == enabled
	case MetricsGateDebug:
		return os.Getenv("METRICS_GATE_DEBUG") == enabled
	case TokenManagerDebug:
		return os.Getenv("TOKEN_MANAGER_DEBUG") == enabled
	case MediationDebug:
		return os.Getenv("MEDIATION_DEBUG") == enabled
	case MetricsEnricherDebug:
		return os.Getenv("METRICS_ENRICHER_DEBUG") == enabled
	case ProductCatalogDebug:
		return os.Getenv("PRODUCT_CATALOG_DEBUG") == enabled
	case ChargingDebug:
		return os.Getenv("CHARGING_DEBUG") == enabled
	case ReportDebug:
		return os.Getenv("REPORT_DEBUG") == enabled
	case ForisPluginDebug:
		return os.Getenv("FORIS_PLUGIN_DEBUG") == enabled
	case MonetizationDebug:
		return os.Getenv("MONETIZATION_DEBUG") == enabled
	default:
		return false
	}
}
