/*
Package bridge provides FreeBSD bridge(4) interface management.

A bridge connects multiple network interfaces together at layer 2, allowing
them to communicate as if they were on the same physical network segment.

# Basic Usage

Create and configure a bridge:

	// Create a new bridge
	br, err := bridge.Create()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created bridge: %s\n", br)

	// Add member interfaces
	if err := bridge.AddMember(br, "em0"); err != nil {
		log.Fatal(err)
	}
	if err := bridge.AddMember(br, "em1"); err != nil {
		log.Fatal(err)
	}

	// Bring the bridge up
	if err := bridge.Up(br, true); err != nil {
		log.Fatal(err)
	}

Query bridge information:

	info, err := bridge.Get("bridge0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bridge %s: MTU=%d, Up=%v\n", info.Name, info.MTU, info.Up)
	fmt.Printf("Members: %v\n", info.Members)

Clean up:

	// Remove member
	if err := bridge.DelMember("bridge0", "em0"); err != nil {
		log.Fatal(err)
	}

	// Destroy bridge
	if err := bridge.Destroy("bridge0"); err != nil {
		log.Fatal(err)
	}

# Permissions

All bridge operations require root privileges.

# Idempotency

Operations are idempotent where appropriate:
  - AddMember: Returns nil if member already exists
  - DelMember: Returns nil if member doesn't exist
*/
package bridge
