# qcon #

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/jbowen.dev/qcon)
[![Go Report Card](https://goreportcard.com/badge/github.com/jamesbo13/qcon)](https://goreportcard.com/report/github.com/jamesbo13/qcon)

`qcon` is a Go library implementing the Synology QuickConnect protocol.

## What is QuickConnect? ##

QuickConnect is a service provided by Synology that allows one to access a
registered Synology NAS device using a globally unique *QuickConnect ID*.
QuickConnect will examine all known routes to the device and connect through
the best available means in the following order:

- local access via LAN
- remote access to device connected directly to internet
- remote access via proxy (eg. router with port-forwarding enabled)
- relay host connected to device via an encrypted tunnel

QuickConnect will only use routes that it has tested for connectivity.

QuickConnect prevents the user from having to manually determine which
method is best for accessing their device.

## Installation ##

```bash
go get github.com/jamesbo13/quickconnect
```

## Usage ##

Most use cases can be handled by the `Resolve()` function:

```go
import (
    "context"
    "github.com/jamesbo13/qcon"
)

...

ctx := context.Background()
id := "your-quick-connect-id"

// Resolve will return list of all working routes to device
// expressed as URL strings. The most preferred route will
// be listed first.
urls, err := qcon.Resolve(ctx, id)
if err != nil {
    // handle error
}

// use synoURL for accessing Synology device APIs
synoURL := urls[0]

...
```

More control over the library behavior can be achieved
by using a custom Client:

```go
import (
    "context"
    "crypto/tls"
    "net/http"
    "github.com/jamesbo13/qcon"
)

...

ctx := context.Background()
id := "your-quick-connect-id"

// Set up a custom http.Transport and Client to control
// individual timeouts and provide server name for TLS cert checks
tr := &http.Transport{
    Dial: (&net.Dialer{
        Timeout: 5 * time.Second,           // Timeout for TCP connection
    }).Dial,
    TLSHandshakeTimeout: 5 * time.Second,   // Timeout for TLS handshake
    TLSClientConfig: &tls.Config{
        ServerName: "synology.mydomain.com",    // Use this name for certificate checks
    },                                          // useful when connecting to IP addresses rather than hostnames
}

c := &qcon.Client{
    Client: &http.Client{
        Timeout: 10 * time.Second,  // Timeout for HTTP responses from server
        Transport: tr,
    },
    Timeout: 20 * time.Second,      // Timeout waiting for connectivity checks
}

// Resolve will return list of all working routes to device
// expressed as URL strings. The most preferred route will
// be listed first.
urls, err := c.Resolve(ctx, id)
if err != nil {
    // handle error
}

// use synoURL for accessing Synology device APIs
synoURL := urls[0]

...
```

## Timeouts and Cancellation ##

The standard `Resolve()` function (and `Client.Resolve()` method) imposes a
default 2 second timeout for its connectivity checks to the URLs it is
testing. Any route that does not respond within the two second timeout will
not be considered for connection.

`Resolve()` and other methods all take the standard `context.Context` parameter.
This can be used to cancel any calls to the library from another goroutine.
See the original [Go blog post on context](https://blog.golang.org/context)
for more details.

## License ##

[MIT](https://choosealicense.com/licenses/mit/)
