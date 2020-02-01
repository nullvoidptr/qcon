package qcon

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
)

const pingPath = "/webman/pingpong.cgi?action=cors&quickconnect=true"

// Ping attempts a ping-pong request to the given URL and returns
// an MD5 hash of the ServerID from the response for use in verification.
func (c Client) Ping(ctx context.Context, url string) (string, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+pingPath, nil)
	if err != nil {
		return "", err
	}

	httpClient := c.Client
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrPingFailure
	}

	var jsonResp struct {
		Success bool
		EZID    string
	}

	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return "", err
	}

	if !jsonResp.Success {
		return "", ErrPingFailure
	}

	return jsonResp.EZID, nil
}

func verifyID(id, hash string) bool {
	h := fmt.Sprintf("%x", md5.Sum([]byte(id)))

	return h == hash
}
