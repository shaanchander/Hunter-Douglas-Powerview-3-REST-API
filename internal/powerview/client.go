package powerview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	host string
	http *http.Client
}

type execPayload struct {
	Hex string `json:"hex"`
}

func NewClient(host string) *Client {
	return &Client{
		host: strings.TrimRight(host, "/"),
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) SendSetPosition(selectedShade string, hexPacket string) (int, error) {
	endpoint := fmt.Sprintf("%s/home/shades/exec?shades=%s", c.host, url.QueryEscape(selectedShade))

	payloadBytes, err := json.Marshal(execPayload{Hex: hexPacket})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
