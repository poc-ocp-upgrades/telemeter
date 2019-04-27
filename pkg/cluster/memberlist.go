package cluster

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"
	"github.com/hashicorp/memberlist"
)

type delegate interface {
	memberlist.EventDelegate
	memberlist.Delegate
}

func NewMemberlist(name, addr string, secret []byte, verbose bool, d delegate) (*memberlist.Memberlist, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(secret) != 32 {
		return nil, fmt.Errorf("invalid secret size, must be 32 bytes: %d", len(secret))
	}
	host, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("address must be a host:port: %v", err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, fmt.Errorf("address must be a host:port: %v", err)
	}
	cfg := memberlist.DefaultWANConfig()
	cfg.DelegateProtocolVersion = protocolVersion
	cfg.DelegateProtocolMax = protocolVersion
	cfg.DelegateProtocolMin = protocolVersion
	cfg.TCPTimeout = 10 * time.Second
	cfg.BindAddr = host
	cfg.BindPort = port
	cfg.AdvertisePort = port
	if !verbose {
		cfg.LogOutput = ioutil.Discard
	}
	cfg.SecretKey = secret
	cfg.Name = name
	cfg.Events = d
	cfg.Delegate = d
	return memberlist.Create(cfg)
}
