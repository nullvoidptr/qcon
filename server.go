package qcon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// fetching and decoding responses from http://global.quickconnect.to/Serv.php

// JSON response
type serverInfo struct {
	Command string
	// Env     json.RawMessage
	ErrNo   int
	Service service
	Server  server
}

// Service
type service struct {
	Port    int
	ExtPort int `json:"ext_port"`
	// PingPong     string
	// PingPongDesc json.RawMessage

	// HTTPS Tunnel Info (not always present)
	RelayIP   string `json:"relay_ip"`
	RelayIPv6 string `json:"relay_ipv6"`
	// RelayDualStack string `json:"relay_dualstack"`
	// RelayDn string `json:"relay_dn"`
	RelayPort int    `json:"relay_port"`
	HttpsIP   string `json:"https_ip"`
	HttpsPort int    `json:"https_port"`
}

// Server info
type server struct {
	DDNS string
	FQDN string
	// Gateway   string
	External  extIPs
	Interface []iface

	ServerID string `json:"serverID"`
	// TCPPunchPort  int    `json:"tcp_punch_port"`
	// UUDPPunchPort int    `json:"udp_punch_port"`
	// Version       string
}

type extIPs struct {
	IP   string
	IPv6 string
}

// Server interface
type iface struct {
	IP   string
	IPv6 []ipv6
	// IPv6Tunnel    []json.RawMessage `json:"ipv6_tunnel"`
	// Mask string
	// Name string
}

// IPv6 address
type ipv6 struct {
	// AddrType     int `json:"addr_type"`
	Address string
	// PrefixLength int `json:"prefix_length"`
	Scope string
}

// commands are either 'get_server_info' or 'request_tunnel'
// ids are 'dsm_portal_https', 'dsm_portal', 'photo_portal_https' or 'photo_portal_http'

const serverQuery = `[
  {
    "version": 1,
    "command": "%s",
    "stop_when_error": false,
    "stop_when_success": false,
    "id": "%s",
    "serverID": "%s",
    "is_gofile": false
  },
  {
    "version": 1,
    "command": "%s",
    "stop_when_error": false,
    "stop_when_success": false,
    "id": "%s",
    "serverID": "%s",
    "is_gofile": false
  }
]`

// QueryURL is URL for global QuickConnect configuration server
//var QueryURL string = "http://global.quickconnect.to/Serv.php"

// newRequestBody returns the JSON encoded body for a request to the server
func newRequestBody(cmd, typ, serverID string) (*bytes.Buffer, error) {

	var b *bytes.Buffer

	if cmd != "get_server_info" && cmd != "request_tunnel" {
		return nil, ErrUnknownCommand
	}

	// TODO: validate serverID

	switch typ {
	case "dsm":
		b = bytes.NewBufferString(fmt.Sprintf(serverQuery, cmd, "dsm_portal_https", serverID, cmd, "dsm_portal", serverID))

	case "photo":
		b = bytes.NewBufferString(fmt.Sprintf(serverQuery, cmd, "photo_portal_https", serverID, cmd, "photo_portal_http", serverID))

	default:
		return nil, ErrUnknownServerType
	}

	return b, nil
}

func getServerInfo(ctx context.Context, c *http.Client, servURL, id string) ([]serverInfo, error) {

	reqBody, err := newRequestBody("get_server_info", "dsm", id)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, servURL, reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info []serverInfo

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}

	if len(info) != 2 {
		return nil, ErrParse
	}

	if info[0].ErrNo != 0 {
		return nil, fmt.Errorf("get_server_info returned errno=%d", info[0].ErrNo)
	}

	if info[1].ErrNo != 0 {
		return nil, fmt.Errorf("get_server_info returned errno=%d", info[1].ErrNo)
	}

	return info, nil
}
