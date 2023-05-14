package arphttp

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"xarpd/arp"
)

type Server struct {
	arpClient *arp.Client
	addr      string
}

func NewServer(arpClient *arp.Client, iface *net.Interface) (*Server, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	if len(addrs) < 1 {
		return nil, fmt.Errorf("no address available for interface %s", iface.Name)
	}
	prefix := netip.MustParsePrefix(addrs[0].String())
	return &Server{
		arpClient: arpClient,
		addr:      fmt.Sprintf("%s:%d", prefix.Addr(), port),
	}, nil
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		switch r.URL.Path {
		case resolvePath:
			s.resolve(w, r)
			return
		case selfPath:
			s.self(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintln(w, "bad request")
}

func (s *Server) resolve(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query().Get("ip")
	ip, err := netip.ParseAddr(param)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "invalid 'ip' query param")
		return
	}

	arpReplies := make(chan net.HardwareAddr)
	s.arpClient.Subscribe(ip, arpReplies)
	defer s.arpClient.Unsubscribe(ip, arpReplies)

	s.arpClient.Request(ip)

	select {
	case mac := <-arpReplies:
		fmt.Fprintln(w, mac)
	case <-r.Context().Done():
	}
}

func (s *Server) self(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, s.arpClient.HardwareAddr())
}
