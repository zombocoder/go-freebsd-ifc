//go:build freebsd
// +build freebsd

package bridge

import (
	"fmt"
	"strings"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
)

// Info represents bridge interface information
type Info struct {
	Name    string
	Members []string
	MTU     int
	Up      bool
}

// MemberFlag is a set of bridge member flags (IFBIF_*), as shown by
// "ifconfig bridgeN" in the "member: X flags=...<...>" lines.
type MemberFlag uint32

const (
	FlagLearning     MemberFlag = constants.IFBIF_LEARNING      // member learns source addresses
	FlagDiscover     MemberFlag = constants.IFBIF_DISCOVER      // member floods packets with unknown destination
	FlagSTP          MemberFlag = constants.IFBIF_STP           // member takes part in spanning tree
	FlagSpan         MemberFlag = constants.IFBIF_SPAN          // member is a span port (read-only)
	FlagSticky       MemberFlag = constants.IFBIF_STICKY        // learned addresses stick
	FlagBSTPEdge     MemberFlag = constants.IFBIF_BSTP_EDGE     // STP edge port
	FlagBSTPAutoEdge MemberFlag = constants.IFBIF_BSTP_AUTOEDGE // STP autoedge enabled
	FlagBSTPPTP      MemberFlag = constants.IFBIF_BSTP_PTP      // STP point-to-point link
	FlagBSTPAutoPTP  MemberFlag = constants.IFBIF_BSTP_AUTOPTP  // STP autoptp enabled
	FlagBSTPAdmEdge  MemberFlag = constants.IFBIF_BSTP_ADMEDGE  // STP admin edge enabled
	FlagBSTPAdmCost  MemberFlag = constants.IFBIF_BSTP_ADMCOST  // STP admin path cost
	FlagPrivate      MemberFlag = constants.IFBIF_PRIVATE       // private segment: no traffic to other private members
	FlagQinQ         MemberFlag = constants.IFBIF_QINQ          // member allows 802.1ad Q-in-Q
)

var memberFlagNames = []struct {
	flag MemberFlag
	name string
}{
	{FlagLearning, "LEARNING"},
	{FlagDiscover, "DISCOVER"},
	{FlagSTP, "STP"},
	{FlagSpan, "SPAN"},
	{FlagSticky, "STICKY"},
	{FlagBSTPEdge, "EDGE"},
	{FlagBSTPAutoEdge, "AUTOEDGE"},
	{FlagBSTPPTP, "PTP"},
	{FlagBSTPAutoPTP, "AUTOPTP"},
	{FlagBSTPAdmEdge, "ADMEDGE"},
	{FlagBSTPAdmCost, "ADMCOST"},
	{FlagPrivate, "PRIVATE"},
	{FlagQinQ, "QINQ"},
}

// Has reports whether every flag in want is set.
func (f MemberFlag) Has(want MemberFlag) bool { return f&want == want }

// String lists the set flags by their ifconfig names, comma separated; bits
// without a name are appended in hex. No bits at all gives "0".
func (f MemberFlag) String() string {
	var parts []string
	rest := f
	for _, n := range memberFlagNames {
		if f&n.flag != 0 {
			parts = append(parts, n.name)
			rest &^= n.flag
		}
	}
	if rest != 0 {
		parts = append(parts, fmt.Sprintf("%#x", uint32(rest)))
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.Join(parts, ",")
}
