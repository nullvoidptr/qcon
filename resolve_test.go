package qcon

import (
	"context"
	"crypto/tls"
	"flag"
	"net/http"
	"testing"
	"time"
)

var testID = flag.String("id", "", "QuickConnect ID")

func Test1(t *testing.T) {

	s := []serverInfo{
		{
			Server: server{
				DDNS:     "test.ddns.org",
				FQDN:     "test.fqdn.com",
				External: extIPs{IP: "11.12.13.14", IPv6: "::"},
				Interface: []iface{
					{
						IP: "192.168.1.2",
						IPv6: []ipv6{
							{Address: "fe80::211:31ff:fe73:cdd7", Scope: "link"},
							{Address: "fd5e:fa9f:15ba:0:211:31ff:fe73:cdd7", Scope: "global"},
						},
					},
					{IP: "10.20.30.40"},
					{IP: "1.2.3.4"},
				},
			},
			Service: service{
				Port:    999,
				ExtPort: 4444,
			},
		},
	}

	s = append(s, s[0])
	s[1].Service.Port = 888
	s[1].Service.ExtPort = 3333

	expected := []string{
		// TODO: httpsSmartLanIPv4
		// TODO: httpsSmartLanIPv6

		// httpsLanIPv4
		"https://192.168.1.2:999",
		"https://10.20.30.40:999",

		// httpsLanIPv6
		"https://[fe80::211:31ff:fe73:cdd7]:999",

		// httpsFQDN
		"https://test.fqdn.com:999",
		"https://test.fqdn.com:4444",

		// httpsDDNS
		"https://test.ddns.org:999",
		"https://test.ddns.org:4444",

		// TODO: httpsSmartHost
		// TODO: httpsSmartWanIPv6
		// TODO: httpsSmartWanIPv4

		// httpsWanIPv6
		"https://[fd5e:fa9f:15ba:0:211:31ff:fe73:cdd7]:999",
		"https://[fd5e:fa9f:15ba:0:211:31ff:fe73:cdd7]:4444",

		// httpsWanIPv4
		"https://1.2.3.4:999",
		"https://11.12.13.14:999",
		"https://11.12.13.14:4444",

		// TODO: httpSmartLanIPv4
		// TODO: httpSmartLanIPv6

		// httpLanIPv4
		"http://192.168.1.2:888",
		"http://10.20.30.40:888",

		// httpLanIPv6
		"http://[fe80::211:31ff:fe73:cdd7]:888",

		// httpFQDN
		"http://test.fqdn.com:888",
		"http://test.fqdn.com:3333",

		// httpDDNS
		"http://test.ddns.org:888",
		"http://test.ddns.org:3333",

		// TODO: httpSmartHost
		// TODO: httpSmartWanIPv6
		// TODO: httpSmartWanIPv4

		// httpWanIPv6
		"http://[fd5e:fa9f:15ba:0:211:31ff:fe73:cdd7]:888",
		"http://[fd5e:fa9f:15ba:0:211:31ff:fe73:cdd7]:3333",

		// httpWanIPv4
		"http://1.2.3.4:888",
		"http://11.12.13.14:888",
		"http://11.12.13.14:3333",
	}

	var urls []string

	for t := uint8(0); t < maxRecordType; t++ {
		var i serverInfo

		if isHTTPS(t) {
			i = s[0]
		} else {
			i = s[1]
		}

		for _, u := range getURLs(i, t) {
			urls = append(urls, u)
		}
	}

	if len(urls) != len(expected) {
		t.Errorf("urls = '%v'", urls)
		t.Fatalf("wrong number of URLs returned: expected %d, got %d", len(expected), len(urls))
	}

	for i := range expected {
		if urls[i] != expected[i] {
			t.Errorf("unexpected / out of order URL:\n got: %s\n exp: %s", urls[i], expected[i])
		}
	}
}

// TODO: Rewrite to test Info.Add()
/*
func TestAddURL(t *testing.T) {

	in := []resolveResp{
		{8, "aaaaa", nil},
		{4, "bbbbb", nil},
		{0, "ccccc", errors.New("error")},
		{0, "ddddd", nil},
		{2, "eeeee", nil},
		{12, "fffff", nil},
	}

	exp := []string{
		"ddddd",
		"eeeee",
		"bbbbb",
		"aaaaa",
		"fffff",
	}

	var out []resolveResp

	for _, r := range in {
		addURL(&out, r)
	}

	if len(exp) != len(out) {
		t.Fatalf("wrong number of strings returned:\n got: %d\n exp: %d", len(out), len(exp))
	}

	for i := range exp {
		if out[i].url != exp[i] {
			t.Errorf("unexpected / out of order item:\n got: %s\n exp: %s", out[i].url, exp[i])
		}
	}
}
*/

func TestGetServers(t *testing.T) {
	if *testID == "" {
		t.Skip()
	}

	ctx := context.Background()

	c := &http.Client{}

	s, err := getServerInfo(ctx, c, defaultServURL, *testID)
	if err != nil {
		t.Fatalf("getInfo() returned error: %s", err)
	}

	t.Logf("%+v\n", s)
}

// TODO: Have it access a test server or test client instead

func TestGetInfoUpdate(t *testing.T) {
	if *testID == "" {
		t.Skip()
	}

	ctx := context.Background()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	c := Client{}
	c.Client = &http.Client{Transport: tr}

	urls, err := c.GetInfo(ctx, *testID)
	if err != nil {
		t.Fatalf("GetInfo() returned error: %s", err)
	}

	t.Logf("%+v\n", urls)

	err = c.UpdateState(ctx, &urls)

	t.Logf("%+v\n", urls)
}

func TestResolve(t *testing.T) {
	if *testID == "" {
		t.Skip()
	}

	ctx := context.Background()

	urls, err := Resolve(ctx, *testID)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", urls)
}

func runResolveTest(t *testing.T, tr *mockTransport, exp []string) {
	t.Helper()

	ctx := context.Background()

	c := Client{
		Client: &http.Client{
			Transport: tr,
		},
		Timeout: 500 * time.Millisecond,
	}

	urls, err := c.Resolve(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != len(exp) {
		t.Fatalf("unexpected number of returned strings. Expected %d, got %d", len(exp), len(urls))
	}

	for i := range urls {
		if urls[i] != exp[i] {
			t.Errorf("returned string mismatch:\n  exp: '%s'\n  got: '%s'", exp[i], urls[i])
		}
	}
}
func TestResolve01(t *testing.T) {

	// Test default response for get_server_info and have only a subset of URLs respond
	tr := &mockTransport{
		responses: map[string]response{
			defaultServURL:                                                   {Status: 200, Body: testServResp},
			"http://75.66.42.168:5000" + pingPath:                            {Status: 200, Body: testPingSuccess},
			"http://10.20.1.100:5000" + pingPath:                             {Status: 200, Body: testPingSuccess},
			"https://75.66.42.168:5001" + pingPath:                           {Status: 200, Body: testPingSuccess},
			"http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5000" + pingPath:   {Status: 200, Body: testPingSuccess},
			"https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50551" + pingPath: {Status: 200, Body: testPingSuccess},
			"https://10.20.1.100:5001" + pingPath:                            {Status: 200, Body: testPingSuccess},
		},
	}

	exp := []string{
		"https://10.20.1.100:5001",                            // httpsLanIPv4
		"https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50551", // httpsLanIPv6
		"https://75.66.42.168:5001",                           // httpsWanIPv4
		"http://10.20.1.100:5000",                             // httpLanIPv4
		"http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5000",   // httpLanIPv6
		"http://75.66.42.168:5000",                            // httpWanIPv4
	}

	runResolveTest(t, tr, exp)
}

func TestResolve02(t *testing.T) {

	// Test a selection of URLs, some of which return invalid ID hash values
	tr := &mockTransport{
		responses: map[string]response{
			defaultServURL:                                                   {Status: 200, Body: testServResp},
			"http://75.66.42.168:5000" + pingPath:                            {Status: 200, Body: testPingInvalid},
			"http://10.20.1.100:5000" + pingPath:                             {Status: 200, Body: testPingSuccess},
			"https://75.66.42.168:5001" + pingPath:                           {Status: 200, Body: testPingInvalid},
			"http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5000" + pingPath:   {Status: 200, Body: testPingInvalid},
			"https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50551" + pingPath: {Status: 200, Body: testPingInvalid},
			"https://10.20.1.100:5001" + pingPath:                            {Status: 200, Body: testPingSuccess},
		},
	}

	exp := []string{
		"https://10.20.1.100:5001", // httpsLanIPv4
		"http://10.20.1.100:5000",  // httpLanIPv4
	}

	runResolveTest(t, tr, exp)
}

func TestResolve03(t *testing.T) {

	// Test a selection of URLs, some of which take too long to return
	tr := &mockTransport{
		responses: map[string]response{
			defaultServURL:                                                   {Status: 200, Body: testServResp},
			"http://75.66.42.168:5000" + pingPath:                            {Status: 200, Body: testPingSuccess, Delay: 2.0},
			"http://10.20.1.100:5000" + pingPath:                             {Status: 200, Body: testPingSuccess, Delay: 0.1},
			"https://75.66.42.168:5001" + pingPath:                           {Status: 200, Body: testPingSuccess, Delay: 2.0},
			"http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5000" + pingPath:   {Status: 200, Body: testPingSuccess, Delay: 2.0},
			"https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50551" + pingPath: {Status: 200, Body: testPingSuccess, Delay: 2.0},
			"https://10.20.1.100:5001" + pingPath:                            {Status: 200, Body: testPingSuccess, Delay: 0.1},
		},
	}

	exp := []string{
		"https://10.20.1.100:5001", // httpsLanIPv4
		"http://10.20.1.100:5000",  // httpLanIPv4
	}

	runResolveTest(t, tr, exp)
}

func TestResolve04(t *testing.T) {

	// Test a selection of URLs, some of which return unexpected body values and/or status errors
	tr := &mockTransport{
		responses: map[string]response{
			defaultServURL:                                                   {Status: 200, Body: testServResp},
			"http://75.66.42.168:5000" + pingPath:                            {Status: 200, Body: "foobar"},
			"http://10.20.1.100:5000" + pingPath:                             {Status: 200, Body: testPingSuccess, Delay: 0.2},
			"https://75.66.42.168:5001" + pingPath:                           {Status: 200, Body: "deadbeef"},
			"http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5000" + pingPath:   {Status: 404, Body: "<html><body>Error</body></html>"},
			"https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50551" + pingPath: {Status: 200, Body: "hello, world!"},
			"https://10.20.1.100:5001" + pingPath:                            {Status: 200, Body: testPingSuccess, Delay: 0.2},
		},
	}

	exp := []string{
		"https://10.20.1.100:5001", // httpsLanIPv4
		"http://10.20.1.100:5000",  // httpLanIPv4
	}

	runResolveTest(t, tr, exp)
}

func TestResolve05(t *testing.T) {

	// Cancel Resolve before it returns, verify correct error returned.
	tr := &mockTransport{
		responses: map[string]response{
			defaultServURL:                                                   {Status: 200, Body: testServResp},
			"http://75.66.42.168:5000" + pingPath:                            {Status: 200, Body: testPingSuccess, Delay: 10},
			"http://10.20.1.100:5000" + pingPath:                             {Status: 200, Body: testPingSuccess, Delay: 10},
			"https://75.66.42.168:5001" + pingPath:                           {Status: 200, Body: testPingSuccess, Delay: 10},
			"http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5000" + pingPath:   {Status: 200, Body: testPingSuccess, Delay: 10},
			"https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50551" + pingPath: {Status: 200, Body: testPingSuccess, Delay: 10},
			"https://10.20.1.100:5001" + pingPath:                            {Status: 200, Body: testPingSuccess, Delay: 10},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// cancel context after 0.5 sec
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	c := Client{
		Client: &http.Client{
			Transport: tr,
		},
		Timeout: 2 * time.Second,
	}

	_, err := c.Resolve(ctx, "foo")

	if err != ErrCancelled {
		if err == nil {
			t.Fatal("cancelled function returned no error")
		}
		t.Fatalf("incorrect error returned")
	}
}

func TestGetInfo(t *testing.T) {

	ctx := context.Background()

	tr := &mockTransport{
		responses: map[string]response{
			defaultServURL: {Status: 200, Body: testServResp},
		},
	}

	c := Client{
		Client: &http.Client{
			Transport: tr,
		},
		Timeout: 500 * time.Millisecond,
	}

	info, err := c.GetInfo(ctx, "foo")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	exp := Info{
		ServerID: "030344165",
		Records: []Record{
			{URL: "https://10.20.1.100:5001", Type: httpsLanIPv4},
			{URL: "https://[fe80::211:32ff:ef63:bca8]:5001", Type: httpsLanIPv6},
			{URL: "https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50551", Type: httpsWanIPv6},
			{URL: "https://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5001", Type: httpsWanIPv6},
			{URL: "https://[fd5e:fa6f:11df::100]:50551", Type: httpsWanIPv6},
			{URL: "https://[fd5e:fa6f:11df::100]:5001", Type: httpsWanIPv6},
			{URL: "https://75.66.42.168:50551", Type: httpsWanIPv4},
			{URL: "https://75.66.42.168:5001", Type: httpsWanIPv4},
			{URL: "http://10.20.1.100:5000", Type: httpLanIPv4},
			{URL: "http://[fe80::211:32ff:ef63:bca8]:5000", Type: httpLanIPv6},
			{URL: "http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:50550", Type: httpWanIPv6},
			{URL: "http://[fd5e:fa6f:11df:0:211:32ff:ef63:bca8]:5000", Type: httpWanIPv6},
			{URL: "http://[fd5e:fa6f:11df::100]:50550", Type: httpWanIPv6},
			{URL: "http://[fd5e:fa6f:11df::100]:5000", Type: httpWanIPv6},
			{URL: "http://75.66.42.168:50550", Type: httpWanIPv4},
			{URL: "http://75.66.42.168:5000", Type: httpWanIPv4},
		},
	}

	if info.ServerID != exp.ServerID {
		t.Errorf("unexpected ServerID:\n  exp: %s\n  got: %s\n", exp.ServerID, info.ServerID)
	}

	if len(info.Records) != len(exp.Records) {
		t.Fatalf("incorrect number of records returned: expected %d, got %d", len(exp.Records), len(info.Records))
	}

	for i := range info.Records {
		if info.Records[i].URL != exp.Records[i].URL {
			t.Errorf("record %d: unexpected URL:\n  exp: %s\n  got: %s\n", i, exp.Records[i].URL, info.Records[i].URL)
		}

		if info.Records[i].Type != exp.Records[i].Type {
			t.Errorf("record %d: unexpected Type:  exp: %d, got: %d", i, exp.Records[i].Type, info.Records[i].Type)
		}
	}
}
