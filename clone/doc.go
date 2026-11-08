/*
Package clone creates and destroys cloned network interfaces by name.

Many FreeBSD interface drivers (wg, gif, gre, lo, ...) are "cloners": the
kernel creates an interface on demand through SIOCIFCREATE. This package is the
generic entry point for drivers that have no dedicated package, or when the
exact interface name matters.

# Basic Usage

Create an interface with an exact name:

	name, err := clone.Create("wg0")
	if err != nil {
		log.Fatal(err)
	}
	defer clone.Destroy(name)

Pass a bare driver name to let the kernel pick the unit:

	name, err := clone.Create("wg") // e.g. "wg0"

# Module Loading

If the kernel does not know the cloner yet, the matching if_<driver> kernel
module is loaded and the create is retried once, as ifconfig(8) does.

# Errors

Creating a name that already exists returns an error for which
errors.Is(err, syscall.EEXIST) is true.

# Permissions

Creating and destroying interfaces requires root privileges.
*/
package clone
