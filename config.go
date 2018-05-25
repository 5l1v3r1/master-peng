/*
MIT License

Copyright (c) 2018 Star Brilliant

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"fmt"
	"net"

	"github.com/BurntSushi/toml"
)

type config struct {
	Left  configInterface
	Right configInterface

	Bandwidth      configBandwidth
	BufferBloat    configBufferBloat
	Cutoff         configCutoff
	DNSPoison      configDNSPoison
	DPI            configDPI
	FakeTraceroute configFakeTraceroute
	IPFirewall     configIPFirewall
	Latency        configLatency
	Loss           configLoss
}

func loadConfig(configFile string) (*config, error) {
	c := &config{}
	metadata, err := toml.DecodeFile(configFile, c)
	if err != nil {
		return nil, err
	}
	for _, key := range metadata.Undecoded() {
		return nil, &configError{fmt.Sprintf("unknown option %q", key.String())}
	}
	c.Left.parse()
	c.Right.parse()
	c.FakeTraceroute.parse()
	c.IPFirewall.parse()
	return c, nil
}

type configError struct {
	err string
}

func (e *configError) Error() string {
	return e.err
}

type configInterface struct {
	DeviceName   string
	HardwareAddr string
	hardwareAddr net.HardwareAddr `toml:"-"`
	IPv4         string
	ipv4Addr     net.IP     `toml:"-"`
	ipv4Net      *net.IPNet `toml:"-"`
	IPv6         string
	ipv6Addr     net.IP     `toml:"-"`
	ipv6Net      *net.IPNet `toml:"-"`
}

func (c *configInterface) parse() error {
	var err error
	if c.HardwareAddr != "" {
		if c.hardwareAddr, err = net.ParseMAC(c.HardwareAddr); err != nil {
			return err
		}
	}
	if c.IPv4 != "" {
		if c.ipv4Addr, c.ipv4Net, err = net.ParseCIDR(c.IPv4); err != nil {
			return err
		}
		if ipv4Addr := c.ipv4Addr.To4(); ipv4Addr == nil {
			return &configError{fmt.Sprintf("invalid IPv4 %q", c.IPv4)}
		} else {
			c.ipv4Addr = ipv4Addr
		}
	}
	if c.IPv6 != "" {
		if c.ipv6Addr, c.ipv6Net, err = net.ParseCIDR(c.IPv6); err != nil {
			return err
		}
		if ipv6Addr := c.ipv6Addr.To16(); ipv6Addr == nil {
			return &configError{fmt.Sprintf("invalid IPv6 %q", c.IPv6)}
		} else {
			c.ipv6Addr = ipv6Addr
		}
	}
	return nil
}

type configGaussDistribution struct {
	Mean  float64
	Stdev float64
}

type configBandwidth struct {
	Enabled         bool
	LeftToRightMbps float64
	RightToLeftMbps float64
}

type configBufferBloat struct {
	Enabled            bool
	LeftToRightPackets uint
	RightToLeftPackets uint
}

type configCutoff struct {
	Enabled         bool
	LeftToRightLoss float64
	LeftToRightOn   configGaussDistribution
	LeftToRightOff  configGaussDistribution
	RightToLeftLoss float64
	RightToLeftOn   configGaussDistribution
	RightToLeftOff  configGaussDistribution
}

type configDNSPoison struct {
	Enabled bool
	Rules   []configDNSPoisonRule
}

type configDNSPoisonRule struct {
	Zone     string
	Addr     []net.IP
	TTL      uint32
	DropUDP  bool
	ResetTCP bool
}

type configDPI struct {
	Enabled bool
	Rules   []configDPIRule
}

type configDPIRule struct {
	Ports           []uint16
	Keyword         string
	BlockWholeIP    bool
	LeftToRightLoss float64
	RightToLeftLoss float64
	ResetTCP        bool
	CooldownTime    float64
}

type configFakeTraceroute struct {
	Enabled bool
	Rules   []configFakeTracerouteRule
}

func (c *configFakeTraceroute) parse() error {
	for _, rule := range c.Rules {
		if err := rule.parse(); err != nil {
			return err
		}
	}
	return nil
}

type configFakeTracerouteRule struct {
	Address     string
	addressCIDR *net.IPNet `toml:"-"`
	Hops        []configFakeTracerouteHop
	Unreachable bool
	Blackhole   bool
}

func (c *configFakeTracerouteRule) parse() error {
	var err error
	if _, c.addressCIDR, err = net.ParseCIDR(c.Address); err != nil {
		return err
	}
	return nil
}

type configFakeTracerouteHop struct {
	Address net.IP
	RTT     float64
}

type configIPFirewall struct {
	Enabled bool
	Rules   []configIPFirewallRule
}

func (c *configIPFirewall) parse() error {
	for _, rule := range c.Rules {
		if err := rule.parse(); err != nil {
			return err
		}
	}
	return nil
}

type configIPFirewallRule struct {
	Address          string
	addressCIDR      *net.IPNet `toml:"-"`
	L4Protocols      []uint16
	Ports            []uint16
	LeftToRightLoss  float64
	RightToLeftLoss  float64
	SourceToDestLoss float64
	DestToSourceLoss float64
	ResetTCP         bool
}

func (c *configIPFirewallRule) parse() error {
	var err error
	if _, c.addressCIDR, err = net.ParseCIDR(c.Address); err != nil {
		return err
	}
	return nil
}

type configLatency struct {
	Enabled        bool
	MaxBufferSize  uint
	LeftToRight    configGaussDistribution
	RightToLeft    configGaussDistribution
	SimulateTunnel bool
	ExcludeICMP    bool
}

type configLoss struct {
	Enabled     bool
	LeftToRight float64
	RightToLeft float64
	ExcludeICMP bool
}
