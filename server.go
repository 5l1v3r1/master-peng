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
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type server struct {
	confPath string
	conf     unsafe.Pointer
	SIGHUP   chan os.Signal

	Left  configInterface
	Right configInterface

	LeftToRight moduleChain
	RightToLeft moduleChain
}

type moduleChain struct {
	Bandwidth      chan frame
	FakeTraceroute chan frame
	Cutoff         chan frame
	Loss           chan frame
	DNSPoison      chan frame
	DPI            chan frame
	IPFirewall     chan frame
	Latency        chan frame
	BufferBloat    chan frame
	Output         chan frame
}

type moduleLayer struct {
	In   chan frame
	Out  chan frame
	Last chan frame
}

type frame []byte

func newServer(confPath string) (*server, error) {
	s := &server{
		confPath: confPath,
		SIGHUP:   make(chan os.Signal, 1),
	}
	err := s.Reload()
	if err != nil {
		return nil, err
	}

	modTunnel := &moduleTunnel{
		Server: s,
		LeftToRight: moduleLayer{
			In:   nil,
			Out:  s.LeftToRight.Output,
			Last: s.LeftToRight.Output,
		},
		RightToLeft: moduleLayer{
			In:   nil,
			Out:  s.RightToLeft.Output,
			Last: s.RightToLeft.Output,
		},
	}

	go func() {
		for {
			_ = <-s.SIGHUP
			err := s.Reload()
			if err != nil {
				log.Println(err)
				continue
			}
			err = modTunnel.Reload()
			if err != nil {
				log.Println(err)
			}
		}
	}()
	signal.Notify(s.SIGHUP, syscall.SIGHUP)
	return s, nil
}

func (s *server) GetConf() *config {
	return (*config)(atomic.LoadPointer(&s.conf))
}

func (s *server) Reload() error {
	conf, err := loadConfig(s.confPath)
	if err != nil {
		return err
	}
	atomic.StorePointer(&s.conf, unsafe.Pointer(conf))
	return nil
}

func (s *server) Start() error {
	return nil
}
