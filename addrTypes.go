package quickconnect

// TODO: change const to allow bitmap operations
// pass to GetInfo() as a filter for only returning certain types

// List of all host/address types ordered by priority
// Lowest values are most preferred
const (
	httpsSmartLanIPv4 uint8 = iota
	httpsSmartLanIPv6
	httpsLanIPv4
	httpsLanIPv6
	httpsFQDN
	httpsDDNS
	httpsSmartHost
	httpsSmartWanIPv6
	httpsSmartWanIPv4
	httpsWanIPv6
	httpsWanIPv4
	httpLanIPv4
	httpLanIPv6
	httpFQDN
	httpDDNS
	httpWanIPv6
	httpWanIPv4
	httpsTun
	httpTun
	maxRecordType
)

// List of only HTTPS address types
var httpsTypes = []uint8{
	httpsSmartLanIPv4,
	httpsSmartLanIPv6,
	httpsLanIPv4,
	httpsLanIPv6,
	httpsFQDN,
	httpsDDNS,
	httpsSmartHost,
	httpsSmartWanIPv6,
	httpsSmartWanIPv4,
	httpsWanIPv6,
	httpsWanIPv4,
	httpsTun,
}

// List of only HTTP address types
var httpTypes = []uint8{
	httpLanIPv4,
	httpLanIPv6,
	httpFQDN,
	httpDDNS,
	httpWanIPv6,
	httpWanIPv4,
	httpTun,
}

func isHTTPS(t uint8) bool {
	return t < httpLanIPv4 || t == httpsTun
}
