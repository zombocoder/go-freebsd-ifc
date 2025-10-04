//go:build freebsd
// +build freebsd

package lagg

import (
	"fmt"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
	"github.com/zombocoder/go-freebsd-ifc/internal/laggops"
)

// Proto represents a LAGG protocol
type Proto int

const (
	ProtoFailover    Proto = 1 // Failover: primary/backup mode
	ProtoLoadBalance Proto = 2 // LoadBalance: hash-based distribution
	ProtoLACP        Proto = 3 // LACP: IEEE 802.3ad
	ProtoRoundRobin  Proto = 4 // RoundRobin: sequential distribution
	ProtoBroadcast   Proto = 5 // Broadcast: replicate to all ports
)

// String returns the protocol name
func (p Proto) String() string {
	switch p {
	case ProtoFailover:
		return "failover"
	case ProtoLoadBalance:
		return "loadbalance"
	case ProtoLACP:
		return "lacp"
	case ProtoRoundRobin:
		return "roundrobin"
	case ProtoBroadcast:
		return "broadcast"
	default:
		return fmt.Sprintf("unknown(%d)", p)
	}
}

// Info represents LAGG interface information
type Info struct {
	Name  string
	Proto Proto
	Ports []string
	MTU   int
	Up    bool
}

// Create creates a new LAGG interface.
//
// The kernel automatically assigns a name (e.g., "lagg0", "lagg1").
// Requires root privileges.
func Create() (string, error) {
	name, err := cloneops.Create("lagg")
	if err != nil {
		return "", fmt.Errorf("create lagg: %w", err)
	}
	return name, nil
}

// Destroy destroys a LAGG interface.
//
// Requires root privileges.
func Destroy(name string) error {
	if err := cloneops.Destroy(name); err != nil {
		return fmt.Errorf("destroy lagg %s: %w", name, err)
	}
	return nil
}

// Up brings the LAGG interface up or down.
//
// Requires root privileges.
func Up(name string, up bool) error {
	if err := ifops.SetFlags(name, uint32(ifc.FlagUp), up); err != nil {
		if up {
			return fmt.Errorf("bring lagg %s up: %w", name, err)
		}
		return fmt.Errorf("bring lagg %s down: %w", name, err)
	}
	return nil
}

// SetProto sets the LAGG protocol.
//
// Must be called before adding ports. Requires root privileges.
func SetProto(name string, proto Proto) error {
	if err := laggops.SetProto(name, int(proto)); err != nil {
		return fmt.Errorf("set lagg %s protocol to %s: %w", name, proto.String(), err)
	}
	return nil
}

// AddPort adds a port to the LAGG interface.
//
// The port interface will be enslaved to the LAGG.
// Requires root privileges.
func AddPort(lagg, port string) error {
	if err := laggops.AddPort(lagg, port); err != nil {
		return fmt.Errorf("add port %s to lagg %s: %w", port, lagg, err)
	}
	return nil
}

// DelPort removes a port from the LAGG interface.
//
// Requires root privileges.
func DelPort(lagg, port string) error {
	if err := laggops.DelPort(lagg, port); err != nil {
		return fmt.Errorf("remove port %s from lagg %s: %w", port, lagg, err)
	}
	return nil
}

// Get returns LAGG interface information.
func Get(name string) (Info, error) {
	cfg, err := laggops.Get(name)
	if err != nil {
		return Info{}, fmt.Errorf("get lagg %s config: %w", name, err)
	}

	iface, err := ifc.Get(name)
	if err != nil {
		return Info{}, fmt.Errorf("get lagg %s interface: %w", name, err)
	}

	return Info{
		Name:  name,
		Proto: Proto(cfg.Proto),
		Ports: cfg.Ports,
		MTU:   iface.MTU,
		Up:    iface.Flags.IsUp(),
	}, nil
}
