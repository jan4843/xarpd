package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/netip"
	"xarpd/arp"
	"xarpd/arphttp"
)

var (
	resolvers []*arphttp.Client
	iface     *net.Interface
)

func init() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Print(`
Usage:
  xarpd [-i INTERFACE] RESOLVER...

Example:
  Forward local ARP requests for 10.0.0.128-10.0.0.255 to another xarpd instance running on 10.0.0.130:
    xarpd -i eth0 10.0.0.130/25
`[1:])
	}
	ifaceName := flag.String("i", defaultIfaceName(), "")
	flag.Parse()

	for _, resolverIP := range flag.Args() {
		prefix, err := netip.ParsePrefix(resolverIP)
		if err != nil {
			log.Fatalf("invalid resolver %q: %s\n", resolverIP, err)
		}
		resolvers = append(resolvers, arphttp.NewClient(prefix))
	}

	foundIface, err := net.InterfaceByName(*ifaceName)
	if err != nil {
		log.Fatalf("invalid interface %q: %s\n", *ifaceName, err)
	}
	iface = foundIface
}

func main() {
	for _, resolver := range resolvers {
		log.Printf("Forwarding ARP requests for %s to %s\n",
			resolver.ServerPrefix().Masked(),
			resolver.ServerPrefix().Addr())
	}

	arpClient, err := arp.NewClient(iface, shouldReplyToARPRequest)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Listening for ARP on %s\n", iface.Name)
	arpClient.Run()

	httpResolver, err := arphttp.NewServer(arpClient, iface)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Listening for HTTP on %s\n", httpResolver.Addr())
	log.Fatalln(httpResolver.Start())
}

func shouldReplyToARPRequest(ip netip.Addr) bool {
	_, err := arphttp.Resolve(ip, resolvers...)
	return err == nil
}
