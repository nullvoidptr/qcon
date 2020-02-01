package qcon

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Default client settings
const (
	defaultTimeout = time.Second * 2
	defaultServURL = "http://global.quickconnect.to/Serv.php"
)

// Client provides HTTP(S) connectivity to the central QuickConnect
// server and is also used for testing connectivity to returned URLs.
//
// Client allows setting of non-default timeout and http.Client options
// for increased control over method behavior. Timeout is separate
// from any timeout settings inherited from http.Client and refers to
// the time Resolve() and UpdateState() will wait for responses during
// connectivity tests.
type Client struct {
	Client  *http.Client
	Timeout time.Duration
	servURL string // Not exported. Only override for testing
}

// DefaultClient is the default Client used by Resolve.
var DefaultClient = &Client{}

// GetInfo returns information for given QuickConnect ID retrieved
// from the global QuickConnect server. This information includes
// the set of all Records associated with this ID (see struct Info).
func (c Client) GetInfo(ctx context.Context, id string) (Info, error) {

	rs := Info{}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// TODO: Handle timeout

	httpClient := c.Client
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	servURL := c.servURL
	if servURL == "" {
		servURL = defaultServURL
	}

	// fetch info on servers
	info, err := getServerInfo(ctx, httpClient, servURL, id)
	if err != nil {
		return rs, err
	}

	if len(info) != 2 {
		return rs, ErrParse
	}

	rs.ServerID = info[0].Server.ServerID

	rs.Records = make([]Record, 0, 16)

	for t := uint8(0); t < maxRecordType; t++ {
		var i serverInfo

		if isHTTPS(t) {
			i = info[0]
		} else {
			i = info[1]
		}

		for _, u := range getURLs(i, t) {
			rs.add(Record{URL: u, Type: t})
		}
	}

	return rs, nil
}

// UpdateState attempts to connect to each URL within Info.Records
// and updates the state value for each. By default, it has a 2 second
// timeout unless Client.Timeout is set to a non-zero value.
func (c Client) UpdateState(ctx context.Context, info *Info) error {

	var err error
	var wg sync.WaitGroup
	var timeout *time.Timer

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	httpClient := c.Client
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	if c.Timeout > 0 {
		timeout = time.NewTimer(c.Timeout)
	} else {
		timeout = time.NewTimer(defaultTimeout)
	}

	ch := make(chan Record)

	for _, r := range info.Records {
		// launch a goroutine to ping each URL from record
		wg.Add(1)
		go func(r Record) {
			defer wg.Done()

			hash, err := c.Ping(ctx, r.URL)
			if err != nil {
				// Need to ensure we simply return if cancelled rather than
				// try to write to channel
				if ctx.Err() != nil {
					return
				}
				r.State = StateConnectFailed
				ch <- r
				return
			}

			// verify ID
			if !verifyID(info.ServerID, hash) {
				r.State = StateInvalidServer
				ch <- r
				return
			}

			r.State = StateOK
			ch <- r

		}(r)

		// Handle cancellation or timeout
		select {
		case <-timeout.C:
			err = ErrTimeout
		case <-ctx.Done():
			err = ErrCancelled
		default:
			break
		}

		if err != nil {
			break
		}

	}

	for err == nil {
		select {
		case r := <-ch:
			for i := range info.Records {
				if r.URL == info.Records[i].URL {
					info.Records[i].State = r.State
					break
				}
			}

		case <-timeout.C:
			err = ErrTimeout
		case <-ctx.Done():
			err = ErrCancelled
		}
	}

	cancel()
	wg.Wait()

	if err == ErrTimeout {
		return nil
	}

	return err
}

func checkExtPort(s serverInfo) bool {
	return s.Service.ExtPort != 0 && s.Service.ExtPort != s.Service.Port
}

func getURLs(s serverInfo, typ uint8) []string {

	var urls []string
	var proto string

	if typ < httpLanIPv4 || typ == httpsTun {
		proto = "https"
	} else {
		proto = "http"
	}

	switch typ {
	// case httpsSmartLanIPv4:
	// case httpsSmartLanIPv6:

	case httpsLanIPv4, httpLanIPv4:

		for _, ifc := range s.Server.Interface {
			if ifc.IP == "" || ifc.IP == "NULL" {
				continue
			}

			if isLocalIP(ifc.IP) {
				urls = append(urls, fmt.Sprintf("%s://%s:%d", proto, ifc.IP, s.Service.Port))
			}
		}

	case httpsWanIPv4, httpWanIPv4:

		for _, ifc := range s.Server.Interface {
			if ifc.IP == "" || ifc.IP == "NULL" {
				continue
			}

			if !isLocalIP(ifc.IP) {
				urls = append(urls, fmt.Sprintf("%s://%s:%d", proto, ifc.IP, s.Service.Port))
			}
		}

		if s.Server.External.IP != "" && s.Server.External.IP != "NULL" && !isLocalIP(s.Server.External.IP) {
			urls = append(urls, fmt.Sprintf("%s://%s:%d", proto, s.Server.External.IP, s.Service.Port))

			if checkExtPort(s) {
				urls = append(urls, fmt.Sprintf("%s://%s:%d", proto, s.Server.External.IP, s.Service.ExtPort))
			}
		}

	case httpsLanIPv6, httpLanIPv6:
		for _, ifc := range s.Server.Interface {
			if len(ifc.IPv6) == 0 {
				continue
			}

			for _, ip := range ifc.IPv6 {
				if ip.Scope == "link" {
					urls = append(urls, fmt.Sprintf("%s://[%s]:%d", proto, ip.Address, s.Service.Port))
				}
			}
		}

	case httpsWanIPv6, httpWanIPv6:
		for _, ifc := range s.Server.Interface {
			if len(ifc.IPv6) == 0 {
				continue
			}

			for _, ip := range ifc.IPv6 {
				if ip.Scope != "link" {
					urls = append(urls, fmt.Sprintf("%s://[%s]:%d", proto, ip.Address, s.Service.Port))
					if checkExtPort(s) {
						urls = append(urls, fmt.Sprintf("%s://[%s]:%d", proto, ip.Address, s.Service.ExtPort))
					}
				}
			}
		}

	case httpsFQDN, httpFQDN:
		if s.Server.FQDN == "" || s.Server.FQDN == "NULL" {
			break
		}

		if checkExtPort(s) {
			urls = make([]string, 2)
			urls[1] = fmt.Sprintf("%s://%s:%d", proto, s.Server.FQDN, s.Service.ExtPort)
		} else {
			urls = make([]string, 1)
		}

		urls[0] = fmt.Sprintf("%s://%s:%d", proto, s.Server.FQDN, s.Service.Port)

	case httpsDDNS, httpDDNS:
		if s.Server.DDNS == "" || s.Server.DDNS == "NULL" {
			break
		}

		if checkExtPort(s) {
			urls = make([]string, 2)
			urls[1] = fmt.Sprintf("%s://%s:%d", proto, s.Server.DDNS, s.Service.ExtPort)
		} else {
			urls = make([]string, 1)
		}

		urls[0] = fmt.Sprintf("%s://%s:%d", proto, s.Server.DDNS, s.Service.Port)

		// case httpsSmartHost:
		// case httpsSmartWanIPv6:
		// case httpsSmartWanIPv4:
		// case httpsWanIPv6:
		// case httpsWanIPv4:
		// case httpLanIPv6:
		// case httpWanIPv6:
		// case httpWanIPv4:
		// case httpsTun:
		// case httpTun:
	}

	return urls
}
