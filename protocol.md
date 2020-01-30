# QuickConnect Protocol #

## Overview ##

QuickConnect is a method to find the most appropriate means of accessing a
Synology NAS device. Given a unique *QuickConnect ID*, one can access the
NAS either locally on LAN, WiFi or remotely over internet. If one is
on the same local network as the Synology device, QuickConnect will provide the
IP information to connect over the local network; if remote, QuickConnect
will handle setting up necessary tunnels to provide secure remote access.

## QuickConnect Process ##

The general process QuickConnect takes to determine the proper connection
mechanism is as follows:

1. Send a `get_server_info` request to the central server at
   `http://global.quickconnect.to/Serv.php` to get information
   for a given QuickConnect ID. Returned information includes
   a list of known IP addresses and/or hostnames for the specified
   device, along with appropriate port numbers and other necessary
   information.

2. Parse the information returned above into a prioritized list of candidate URLs that may
connect to the desired device (depending on the
network connectivity of the client device).

3. Loop through all IP addresses and standard DSM port addresses to
   find a response on http(s)://<IP>:<port>/webman/pingpong.cgi.

4. From those IP addresses that respond, select one and use that
   for API requests.

5. If no hosts respond correctly, the device is not accessible locally
   or directly over the internet. In this case, QuickConnect makes a
   `request_tunnel` request to http://global.quickconnect.to/Serv.php.
   This will return a relay IP address and port that can be used to
   connect to the NAS device.

These steps are outlined in further detail below.

### Step 1: `get_server_info` ###

The first operation in the QuickConnect mechanism is to make a `get_server_info`
request to the central QuickConnect server.

#### Request ####

`POST http://global.quickconnect.to/Serv.php`

with the following form data:

```json
[
  {
    "version": 1,
    "command": "get_server_info",
    "stop_when_error": false,
    "stop_when_success": false,
    "id": "dsm_portal_https",
    "serverID": "<QUICKCONNECT_ID>",
    "is_gofile": false
  },
  {
    "version": 1,
    "command": "get_server_info",
    "stop_when_error": false,
    "stop_when_success": false,
    "id": "dsm_portal",
    "serverID": "<QUICKCONNECT_ID>",
    "is_gofile": false
  }
]
```

where `<QUICKCONNECT_ID>` is replaced with the desired QuickConnect ID.

#### Response ####

The server returns a JSON-encoded response similar to the following:

```json
[
  {
    "command": "get_server_info",
    "env": {
      "control_host": "usc.quickconnect.to",
      "relay_region": "us"
    },
    "errno": 0,
    "server": {
      "ddns": "NULL",
      "ds_state": "CONNECTED",
      "external": {
        "ip": "<EXTERNAL_IP>",
        "ipv6": "::"
      },
      "fqdn": "NULL",
      "gateway": "<GATEWAY_IP>",
      "interface": [
        {
          "ip": "<LOCAL_IP",
          "ipv6": [
            {
              "addr_type": 32,
              "address": "<IPV6_LINK_LOCAL>",
              "prefix_length": 64,
              "scope": "link"
            },
            {
              "addr_type": 0,
              "address": "<IPV6_GLOBAL1>",
              "prefix_length": 64,
              "scope": "global"
            },
            {
              "addr_type": 0,
              "address": "<IPV6_GLOBAL2>",
              "prefix_length": 64,
              "scope": "global"
            }
          ],
          "mask": "255.255.255.0",
          "name": "eth0"
        }
      ],
      "ipv6_tunnel": [],
      "serverID": "<SERVER_ID>",
      "tcp_punch_port": 0,
      "udp_punch_port": 36810,
      "version": "24922"
    },
    "service": {
      "port": 5001,
      "ext_port": 50551,
      "pingpong": "DISCONNECTED",
      "pingpong_desc": []
    },
    "version": 1
  },
  {
    "command": "get_server_info",
    "env": {
      "control_host": "usc.quickconnect.to",
      "relay_region": "us"
    },
    "errno": 0,
    "server": {
      "ddns": "NULL",
      "ds_state": "CONNECTED",
      "external": {
        "ip": "<EXTERNAL_IP>",
        "ipv6": "::"
      },
      "fqdn": "NULL",
      "gateway": "<GATEWAY_IP>",
      "interface": [
        {
          "ip": "<LOCAL_IP",
          "ipv6": [
            {
              "addr_type": 32,
              "address": "<IPV6_LINK_LOCAL>",
              "prefix_length": 64,
              "scope": "link"
            },
            {
              "addr_type": 0,
              "address": "<IPV6_GLOBAL1>",
              "prefix_length": 64,
              "scope": "global"
            },
            {
              "addr_type": 0,
              "address": "<IPV6_GLOBAL2>",
              "prefix_length": 64,
              "scope": "global"
            }
          ],
          "mask": "255.255.255.0",
          "name": "eth0"
        }
      ],
      "ipv6_tunnel": [],
      "serverID": "<SERVER_ID>",
      "tcp_punch_port": 0,
      "udp_punch_port": 36810,
      "version": "24922"
    },
    "service": {
      "port": 5000,
      "ext_port": 50550,
      "pingpong": "DISCONNECTED",
      "pingpong_desc": []
    },
    "version": 1
  }
]
```

where `<GATEWAY_IP>`, `<LOCAL_IP>`, `<IPV6_LINK_LOCAL>`, `<IPV6_GLOBAL1>` and `<IPV6_GLOBAL2>`
are device specific network settings.

As one can see there are two separate responses (correlating to our two initial requests),
one for HTTP and one for HTTPS traffic. While the IP addresses are the same in this case,
the port numbers will differ.

In addition to the various hostnames, IP addresses and port numbers, there is a key field
in the response called `serverID` which is used to validate "ping-pong" responses in
the steps that follow.

### Step 2: Parse Server Info into URL Candidates ###

After retrieving the server information, QuickConnect will process the JSON response into
a series of URLs that will have a simple HTTP/HTTPS "ping-pong" request sent to test connectivity
to that address/host.

In prioritized order (most favored first), the URLs considered are as follows:

Protocol|Description
--------|-----------
HTTPS|Smart DNS IPv4 Local Network (LAN)
HTTPS|Smart DNS IPv6 Local Network (LAN)
HTTPS|IPv4 Local Network (LAN)
HTTPS|IPv6 Local Network (LAN)
HTTPS|Fully Qualified Domain Name (FQDN)
HTTPS|Dynamic DNS Name (DDNS)
HTTPS|Smart DNS Host
HTTPS|Smart DNS IPv6 Remote Network (WAN)
HTTPS|Smart DNS IPv4 Remote Network (WAN)
HTTPS|IPv6 Remote Network (WAN)
HTTPS|IPv4 Remote Network (WAN)
HTTP|IPv4 Local Network (LAN)
HTTP|IPv6 Local Network (LAN)
HTTP|Fully Qualified Domain Name (FQDN)
HTTP|Dynamic DNS Name (DDNS)
HTTPS|IPv6 Remote Network (WAN)
HTTPS|IPv4 Remote Network (WAN)
HTTPS|QuickConnect Tunnel
HTTP|QuickConnect Tunnel

As one can see, with the exception of QuickConnect Tunnels (a last resort if the device is not 
otherwise accessible), HTTPS URLs are prioritized over HTTP URLs. Likewise, local network access
is preferable to connecting over the internet or
via a tunnel. Finally, within local networks, IPv4
is preferred over IPv6, but the opposite applies on
remote networks where IPv6 is preferred.

Details of how each URL is generated follow below:

#### Smart DNS IPv4 (WAN/LAN) ####

If the server info response contains a `smartdns` object in the root
response object (ie. as a peer to `server` and `service`), any values
contained in `smartdns:lan` array that do NOT start with `syn4-` will
be treated as LAN addresses, otherwise they will be WAN addresses.

LAN types will only have service port checked (`service:port` in JSON), but WAN types
will have `service:ext_port` checked as well.

### Smart DNS IPv6 (WAN/LAN) ###

Similar to above, any values in the `smartdns:lanv6` array (if present)
that do NOT start with `syn6-` will be treated as LAN addresses, otherwise they will be categorized as WAN addresses.

LAN types will only have service port checked (`service:port` in JSON), but WAN types
will have `service:ext_port` checked as well.

### IPv4 Addresses (WAN/LAN) ###

Within the `server` object in the JSON response, there is an
`interface` array. Each member of this array can contain an
`ip` field containing a string of a single IPv4 address.

The IP address is checked to see if it is a loopback or private
address (based on RFC 1918 IP assignment rules). This will
determine if the IP is treated as LAN (private/loopback) or 
WAN (any other IP address).

LAN types will only have service port checked (`service:port` in JSON), but WAN types
will have `service:ext_port` checked as well.

In addition, if the field `server:external:ip` is defined, it will be used
as an IPv4 WAN address.

### IPv6 Addresses (WAN/LAN) ###

Within the `server` object in the JSON response, there is an
`interface` array. Each member of this array can contain an
`ipv6` array containing zero or more IPv6 address objects.

For each IPv6 object discovered,  the `scope` field of the
object is checked to determine whether it is a local or remote
address. If `scope` == "link", the address is categorized as
local (LAN). Otherwise it is a remote address (WAN).

"lan" types will only have service port checked (`service:port` in JSON), but "wan" types \
will have `service:ext_port` checked as well.

LAN types will only have service port checked (`service:port` in JSON), but WAN types
will have `service:ext_port` checked as well.

### Fully Qualified Domain Name (FQDN) ###

If the field `server:fqdn` is present and not empty, it will be used. Both
`service:port` and `service:ext_port` will be checked.

### Dynamic DNS Name (DDNS) ###

If the field `server:ddns` is present and not empty, it will be used. Both
`service:port` and `service:ext_port` will be checked.

### Smart DNS Host ###

In addition to the `lan` and `lan6` arrays within the `smartdns` object,
there is also the possibility of a string field named `host`. If present,
this value is categorized as a Smart DNS hostname both `service:port` and `service:ext_port` checked.

### Note on service and ext_port checks ###

External ports only checked for some types (see above) and only if the
returned value for `service.ext_port` is not empty, "0" or equal to
`service.port`.

### Step 3: Test Network Accessibility ###

For each of the addresses parsed into URLs in the previous step,
QuickConnect will attempt to connect to the host using a predefined
"ping-pong" URL. If a successful and validated response is returned,
the URL is added to the list of accessible candidates.

#### Request ####

For address being tested, an HTTP (or HTTPS) query is sent to
`http(s)://<address>:<port>/webman/pingpong.cgi` with the query parameters,
`action=cors&quickconnect=true`.

#### Response ####
If the host is a Synology NAS device running DSM it will return
a JSON encoded response:

```json
{
  "success": true,
  "ezid": "d98a13037fbfebb9f5ce438cf9634050"
}
```

Any queries to a host that is not accessible will likely simply not respond,
return an HTTP error code (eg. 4040) or simply return unexpected content.

#### Validation ####

In addition to verifying the HTTP status code (200) and the response
is proper JSON in the desired format, QuickConnect also examines the
`ezid` field in the response and uses it to verify the host is not
just a Synology NAS device but the precise device we expect.

The `ezid` parameter is an MD5 hash of the `server:serverID` value returned
in the `get_server_info` response. If `ezid` does not match the
MD5 hash of the `serverID`, the candidate host will be ignored.

#### Timeouts ####

It is expected that not all "ping-pong" queries will respond. As such
it is necessary for QuickConnect to set a timeout on waiting for
responses. Examining the behavior of the standard javascript QuickConnect
client, it appears the default timeout is 2 seconds from the time of
the first "ping-pong" query.

### Step 4: Prioritize Successful Responses ###

As mentioned in Step 2, each generated URL is categorized as a particular
type with a particular priority. Once timeout in Step 3 has been reached,
all the successful and validated responses are checked to find the URL
with the highest priority. Within the javascript QuickConnect client,
the current browser instance will be reloaded and redirected to this
URL.

If there are no successful responses to Step 3, then it means:

- the NAS device is on a different network than the client (ie. LAN addresses are inaccessible)

- firewall rules are preventing direct access to the NAS device over
the internet

In this case, we need to set up a QuickConnect tunnel (Step 5 below).

### Step 5: Establish Tunnel If Necessary ###

When QuickConnect is enabled on a Synology NAS, it runs the `synorelayd`
daemon. This daemon connects to a server on the internet run
by Synology that provides an always-open, encrypted, two-way connection to
the NAS device. This connection, allows the remote server to request
an OpenVPN tunnel to be created when necessary even if the NAS
device is not otherwise accessible from the internet.

To setup a tunnel, the QuickConnect client sends a `request_tunnel`
request to the same main URL as in Step 1. The server will then
respond with a similar JSON response as in Step 1. The only difference
will be the presence of additional fields in the `service` object:

```json
      "relay_ip": "89.187.18.191",
      "relay_ipv6": "2b02:9df0:c80d::84",
      "relay_dualstack": "usrds5.xxxx.yyyy.quickconnect.to",
      "relay_dn": "usr5.xxxx.yyyy.quickconnect.to",
      "relay_port": 2905,
      "https_ip": "89.187.18.191",
      "https_port": 443
```

The `relay_ip` (or `relay_ipv6`) and `relay_port` can be used to connect
to the Synology device via the established tunnel. It is currently unclear
of the significance of the other new fields.

If the response is successful and contains the above fields, new URLs
can be formed as follows:

- `https://<relay_ip>:<relay_port>`
- `https://<relay_ipv6>:<relay_port>`
- `http://<relay_ip>:<relay_port>`
- `http://<relay_ipv6>:<relay_port>`

Connectivity to these URLs will be tested in the same means as before
("ping-pong" request). HTTPS connections will be prioritized over HTTP
but it is unclear whether IPv4 or IPv6 is preferred over the other. Given
the prioritization of IPv6 over IPv4 for WAN traffic, I would assume
the same prioritization would occur in this case but I have not tested
to verify. 

### Notes on HTTPS Certificate Validation ###

Proper HTTPS traffic requires validation of the HTTPS server's identity
to ensure againt "man-in-the-middle" attacks. Due to the design of HTTPS
and the underlying TLS encryption, certificates cannot be assigned to or
validated against an IP address. One option to access HTTPS URLs that
have IP address (rather than domain names) is to disable certificate
checking. This, however, **IS VERY INSECURE** and defeats one of the
main security features of HTTPS.

Instead you need to ensure the domain name assigned to the certificate
on the Synology device resolves to the desired IP address. This won't
be the case for Tunnels or (likely) local addresses. Instead, one
needs to override the DNS resolver to associate the IP address with
the certificate name.

In curl this can be done as follows:
```
curl -v https://syno.mydomain.com:30783 --resolve "syno.mydomain.com:30783:88.182.193.22"
```

How can we do this in Go? 