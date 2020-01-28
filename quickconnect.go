/*
Package quickconnect implements the QuickConnect protocol for
accessing Synology NAS devices over the best available connection
using a globally unique identifier. The returned URLs will vary
depending on the client network relative to the Synology - for
example, if on the same LAN local URLs will be returned, otherwise
URLs using remote address (or even a Synology created tunnel)
will be returned.

For most cases, all that is necessary is to pass the QuickConnect ID
to the Resolve() function which will return a prioritized list of
URLs:

	// Create context
	ctx := context.Background()

	// Fetch list of URLs
	urls, err := quickconnect.Resolve(ctx, id)

	// Provided no error returned, the first URL in the returned slice
	// will be the most desired available connection. This can then
	// be used to access the Synology API

More control can be obtained by creating a Client with custom
settings (including modifying the default http.Client).

*/
package quickconnect
