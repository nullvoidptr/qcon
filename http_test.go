package quickconnect

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	testServResp    = `[{"command":"get_server_info","env":{"control_host":"usc.quickconnect.to","relay_region":"us"},"errno":0,"server":{"ddns":"NULL","ds_state":"CONNECTED","external":{"ip":"75.66.42.168","ipv6":"::"},"fqdn":"NULL","gateway":"10.20.1.1","interface":[{"ip":"10.20.1.100","ipv6":[{"addr_type":32,"address":"fe80::211:32ff:ef63:bca8","prefix_length":64,"scope":"link"},{"addr_type":0,"address":"fd5e:fa6f:11df::100","prefix_length":64,"scope":"global"},{"addr_type":0,"address":"fd5e:fa6f:11df:0:211:32ff:ef63:bca8","prefix_length":64,"scope":"global"}],"mask":"255.255.255.0","name":"eth0"}],"ipv6_tunnel":[],"serverID":"030344165","tcp_punch_port":0,"udp_punch_port":36810,"version":"24922"},"service":{"port":5001,"ext_port":50551,"pingpong":"DISCONNECTED","pingpong_desc":[]},"version":1},{"command":"get_server_info","env":{"control_host":"usc.quickconnect.to","relay_region":"us"},"errno":0,"server":{"ddns":"NULL","ds_state":"CONNECTED","external":{"ip":"75.66.42.168","ipv6":"::"},"fqdn":"NULL","gateway":"10.20.1.1","interface":[{"ip":"10.20.1.100","ipv6":[{"addr_type":32,"address":"fe80::211:32ff:ef63:bca8","prefix_length":64,"scope":"link"},{"addr_type":0,"address":"fd5e:fa6f:11df::100","prefix_length":64,"scope":"global"},{"addr_type":0,"address":"fd5e:fa6f:11df:0:211:32ff:ef63:bca8","prefix_length":64,"scope":"global"}],"mask":"255.255.255.0","name":"eth0"}],"ipv6_tunnel":[],"serverID":"030344165","tcp_punch_port":0,"udp_punch_port":36810,"version":"24922"},"service":{"port":5000,"ext_port":50550,"pingpong":"DISCONNECTED","pingpong_desc":[]},"version":1}]`
	testPingSuccess = `{"success": true,"ezid": "36e618cde8a29a8a8ef945ae21402312"}`
	testPingFail    = `{"success": false}`
	testPingInvalid = `{"success": true,"ezid": "00000000000000000000000000000000"}`
)

// mockTransport is a drop-in replacement for http.Transport
// which implements the http.RoundTripper interface. It is
// used for mocking HTTP client-server requests without requiring
// a server.
type mockTransport struct {
	responses map[string]response
}

type response struct {
	Status int
	Body   string
	Delay  float32
}

type config map[string]response

// Creates a new MockTransport using the responses defined in file fn
func newMockTransport(fn string) (*mockTransport, error) {

	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	var cfg config

	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &mockTransport{responses: cfg}, nil
}

func (t mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u := req.URL.String()

	// fmt.Printf(" > %s %s\n", req.Method, u)

	resp := &http.Response{
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
	}

	r, ok := t.responses[u]
	if !ok {
		// fmt.Printf(" < ERROR\n")
		return nil, errors.New("unknown URL: no response in config")
	}

	if r.Delay > 0 {
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.NewTimer(time.Duration(r.Delay*1000) * time.Millisecond).C:
			break
		}
	}

	resp.StatusCode = r.Status
	resp.Body = ioutil.NopCloser(strings.NewReader(r.Body))

	// fmt.Printf(" < %s\n", r.Body)

	return resp, nil
}
