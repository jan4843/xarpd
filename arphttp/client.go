package arphttp

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type Client struct {
	serverPrefix netip.Prefix
	httpClient   *http.Client
}

func NewClient(serverPrefix netip.Prefix) *Client {
	return &Client{
		serverPrefix: serverPrefix,
		httpClient:   &http.Client{Timeout: resolveTimeout},
	}
}

func (c *Client) ServerPrefix() netip.Prefix {
	return c.serverPrefix
}

func (c *Client) ResolveWithContext(ctx context.Context, ip netip.Addr) (net.HardwareAddr, error) {
	url := fmt.Sprintf("http://%s:%d%s?ip=%s", c.serverPrefix.Addr(), port, resolvePath, ip)
	if c.serverPrefix.Addr() == ip {
		url = fmt.Sprintf("http://%s:%d%s", c.serverPrefix.Addr(), port, selfPath)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return net.ParseMAC(strings.TrimSpace(string(body)))
}
