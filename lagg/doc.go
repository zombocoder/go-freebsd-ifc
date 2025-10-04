//go:build freebsd
// +build freebsd

/*
Package lagg provides FreeBSD LAGG (Link Aggregation) interface management.

LAGG interfaces combine multiple network interfaces into a single logical interface
for redundancy, load balancing, or increased bandwidth.

# Basic Usage

Create and configure a LAGG interface:

	// Create LAGG interface
	name, err := lagg.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer lagg.Destroy(name)

	// Set protocol to LACP
	if err := lagg.SetProto(name, lagg.ProtoLACP); err != nil {
		log.Fatal(err)
	}

	// Add member ports
	if err := lagg.AddPort(name, "em0"); err != nil {
		log.Fatal(err)
	}
	if err := lagg.AddPort(name, "em1"); err != nil {
		log.Fatal(err)
	}

	// Bring interface up
	if err := lagg.Up(name, true); err != nil {
		log.Fatal(err)
	}

# Protocols

FreeBSD supports several link aggregation protocols:

  - ProtoFailover: Sends traffic through the primary port; fails over to secondary ports
  - ProtoLoadBalance: Balances outgoing traffic across all ports using source/dest hash
  - ProtoLACP: IEEE 802.3ad Link Aggregation Control Protocol (dynamic)
  - ProtoRoundRobin: Distributes outgoing traffic in round-robin fashion
  - ProtoBroadcast: Sends all traffic on all ports simultaneously

# Permissions

All operations require root privileges.

# Common Use Cases

  - Server redundancy: Failover protocol for high availability
  - Network throughput: LACP or LoadBalance for increased bandwidth
  - Switch stacking: Aggregating links between switches

# Example

	// Create LAGG with LACP for redundancy and load balancing
	lagg0, _ := lagg.Create()
	lagg.SetProto(lagg0, lagg.ProtoLACP)
	lagg.AddPort(lagg0, "em0")
	lagg.AddPort(lagg0, "em1")
	lagg.Up(lagg0, true)
*/
package lagg
