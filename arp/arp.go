package arp

import (
	"net"
	"net/netip"
	"sync"

	"github.com/mdlayher/arp"
)

type Client struct {
	client        *arp.Client
	mutex         sync.Mutex
	subs          map[netip.Addr][]chan net.HardwareAddr
	handleRequest RequestHandler
}

type RequestHandler func(ip netip.Addr) (shouldReply bool)

func NewClient(iface *net.Interface, handleRequest RequestHandler) (*Client, error) {
	client, err := arp.Dial(iface)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:        client,
		subs:          make(map[netip.Addr][]chan net.HardwareAddr),
		handleRequest: handleRequest,
	}, nil
}

func (c *Client) HardwareAddr() net.HardwareAddr {
	return c.client.HardwareAddr()
}

func (c *Client) Subscribe(ip netip.Addr, ch chan net.HardwareAddr) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, existingCh := range c.subs[ip] {
		if ch == existingCh {
			return
		}
	}
	c.subs[ip] = append(c.subs[ip], ch)
}

func (c *Client) Unsubscribe(ip netip.Addr, ch chan net.HardwareAddr) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, replyCh := range c.subs[ip] {
		if replyCh == ch {
			c.subs[ip][i] = c.subs[ip][len(c.subs[ip])-1]
			c.subs[ip][len(c.subs[ip])-1] = nil
			c.subs[ip] = c.subs[ip][:len(c.subs[ip])-1]
			break
		}
	}
	if len(c.subs[ip]) == 0 {
		delete(c.subs, ip)
	}
}

func (c *Client) Request(ip netip.Addr) {
	c.client.Request(ip)
}

func (c *Client) Run() {
	go func() {
		for {
			pkt, _, err := c.client.Read()
			if err != nil {
				continue
			}

			switch pkt.Operation {
			case arp.OperationReply:
				go c.handleReply(pkt)
			case arp.OperationRequest:
				go func() {
					shouldReply := c.handleRequest(pkt.TargetIP)
					if shouldReply {
						c.reply(pkt)
					}
				}()
			}
		}
	}()
}

func (c *Client) handleReply(pkt *arp.Packet) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, ch := range c.subs[pkt.SenderIP] {
		select {
		case ch <- pkt.SenderHardwareAddr:
		default:
		}
	}
}

func (c *Client) reply(requestPkt *arp.Packet) error {
	reply, err := arp.NewPacket(arp.OperationReply,
		c.HardwareAddr(), requestPkt.TargetIP,
		requestPkt.SenderHardwareAddr, requestPkt.SenderIP,
	)
	if err != nil {
		return err
	}

	return c.client.WriteTo(reply, requestPkt.SenderHardwareAddr)
}
