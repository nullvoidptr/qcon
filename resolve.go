package quickconnect

import (
	"context"
)

// Resolve returns a list of URL strings for accessing the server
// using the provided QuickConnect ID. The URL strings are in
// ranked order, most preferred first and only those with verified
// connectivity are returned.
//
// Resolve is a wrapper to (Client) Resolve() using DefaultClient
func Resolve(ctx context.Context, id string) ([]string, error) {

	c := DefaultClient
	return c.Resolve(ctx, id)
}

// Resolve returns a list of URL strings for accessing the server
// using the provided QuickConnect ID. The URL strings are in
// ranked order, most preferred first and only those with verified
// connectivity are returned.
func (c Client) Resolve(ctx context.Context, id string) ([]string, error) {

	info, err := c.GetInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	err = c.UpdateState(ctx, &info)
	if err != nil {
		return nil, err
	}

	var urls []string

	for _, r := range info.Records {
		if r.State == StateOK {
			urls = append(urls, r.URL)
		}
	}

	if len(urls) == 0 {
		return nil, ErrCannotAccess
	}

	return urls, nil
}
