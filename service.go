package quickconnect

import (
	"net"
	"sort"
)

// Record is a single QuickConnect redirect record indicating a
// URL that may be able to access the desired Synology service.
// Each record has a Type which is used to prioritize URLs and
// a State to indicate the result of the most recent connection
// test with that host.
type Record struct {
	URL   string
	Type  uint8
	State ConnState
}

// ConnState indicates the connection state with a URL/host
type ConnState uint8

const (
	StateUnknown ConnState = iota
	StateOK
	StateConnectFailed
	StateInvalidServer
)

// Info contains information on a QuickConnect host
type Info struct {
	ServerID string
	Records  []Record
}

// Add Record to Info, sorted by Record.Type
func (set *Info) add(r Record) {

	s := set.Records

	// Find insertion point
	i := sort.Search(len(s), func(i int) bool { return s[i].Type >= r.Type })

	if i == len(s) {
		// append record to end of current set
		s = append(s, r)
	} else {
		// insert at index i

		// expand the slice if necessary using zero value placeholder
		s = append(s, Record{})

		// shift any items at insertion point or after one spot over
		copy(s[i+1:], s[i:])

		// insert new item
		s[i] = r
	}

	set.Records = s
}

// inspired by / copied from https://go-review.googlesource.com/c/go/+/162998/7/src/net/ip.go
func isLocalIP(addr string) bool {

	ip := net.ParseIP(addr)

	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1]&0xf0 == 16) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}

	return len(ip) == net.IPv6len && ip[0]&0xfe == 0xfc
}
