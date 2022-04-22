package ports

// Allocator - аллокатор портов
type Allocator interface {
	NextPort() uint16
}

type portsAllocator struct {
	startPort uint16
}

// NewPortsAllocator - конструктор аллокатора портов
func NewPortsAllocator(start TestPortRange) Allocator {
	return &portsAllocator{startPort: uint16(start)}
}

// NextPort - возвращает следующий незанятый порт
func (pa *portsAllocator) NextPort() uint16 {
	pa.startPort++

	return pa.startPort - 1
}
