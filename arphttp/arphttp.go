package arphttp

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"
)

const (
	port        = 2707
	resolvePath = "/resolve"
	selfPath    = "/self"

	resolveTimeout = 3 * time.Second
)

func Resolve(ip netip.Addr, clients ...*Client) (net.HardwareAddr, error) {
	var result net.HardwareAddr
	resolved := false

	var wg sync.WaitGroup
	var once sync.Once
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, client := range clients {
		client := client
		if !client.ServerPrefix().Contains(ip) {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			mac, err := client.ResolveWithContext(ctx, ip)
			if err != nil {
				return
			}
			once.Do(func() {
				result = mac
				resolved = true
				cancel()
			})
		}()
	}
	wg.Wait()

	if resolved {
		return result, nil
	}
	return nil, fmt.Errorf("%s: cannot resolve", ip)
}
