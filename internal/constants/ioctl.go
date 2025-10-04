//go:build freebsd
// +build freebsd

package constants

/*
#include <sys/sockio.h>
#include <net/if.h>
#include <net/if_bridgevar.h>
#include <net/if_vlan_var.h>
#include <net/if_lagg.h>
#include <netinet/in.h>
#include <netinet6/in6_var.h>
*/
import "C"

// Interface ioctls
const (
	SIOCGIFFLAGS  = C.SIOCGIFFLAGS
	SIOCSIFFLAGS  = C.SIOCSIFFLAGS
	SIOCGIFMTU    = C.SIOCGIFMTU
	SIOCSIFMTU    = C.SIOCSIFMTU
	SIOCSIFNAME   = C.SIOCSIFNAME
	SIOCIFCREATE  = C.SIOCIFCREATE
	SIOCIFDESTROY = C.SIOCIFDESTROY
	SIOCSDRVSPEC  = C.SIOCSDRVSPEC
	SIOCGDRVSPEC  = C.SIOCGDRVSPEC
)

// Bridge ioctls
const (
	BRDGADD  = C.BRDGADD
	BRDGDEL  = C.BRDGDEL
	BRDGGIFS = C.BRDGGIFS
)

// VLAN ioctls
const (
	SIOCSETVLAN = C.SIOCSETVLAN
	SIOCGETVLAN = C.SIOCGETVLAN
)

// LAGG ioctls
const (
	SIOCSLAGG        = C.SIOCSLAGG
	SIOCGLAGG        = C.SIOCGLAGG
	SIOCSLAGGPORT    = C.SIOCSLAGGPORT
	SIOCSLAGGDELPORT = C.SIOCSLAGGDELPORT
	SIOCGLAGGPORT    = C.SIOCGLAGGPORT
)

// IP address ioctls
const (
	SIOCAIFADDR     = C.SIOCAIFADDR
	SIOCDIFADDR     = C.SIOCDIFADDR
	SIOCAIFADDR_IN6 = C.SIOCAIFADDR_IN6
	SIOCDIFADDR_IN6 = C.SIOCDIFADDR_IN6
)

// Interface flags
const (
	IFF_UP          = C.IFF_UP
	IFF_BROADCAST   = C.IFF_BROADCAST
	IFF_DEBUG       = C.IFF_DEBUG
	IFF_LOOPBACK    = C.IFF_LOOPBACK
	IFF_POINTOPOINT = C.IFF_POINTOPOINT
	IFF_RUNNING     = C.IFF_RUNNING
	IFF_NOARP       = C.IFF_NOARP
	IFF_PROMISC     = C.IFF_PROMISC
	IFF_ALLMULTI    = C.IFF_ALLMULTI
	IFF_MULTICAST   = C.IFF_MULTICAST
)

// Address families
const (
	AF_INET  = C.AF_INET
	AF_INET6 = C.AF_INET6
	AF_LINK  = C.AF_LINK
)

// Structure sizes
const (
	IFNAMSIZ          = C.IFNAMSIZ
	SizeofSockaddrIn  = C.sizeof_struct_sockaddr_in
	SizeofSockaddrIn6 = C.sizeof_struct_sockaddr_in6
)

// IPv6 constants
const (
	ND6_INFINITE_LIFETIME = 0xffffffff
)
