package proxy

import (
	"context"
	"fmt"
	"testing"
)

func TestManager_Run(t *testing.T) {
	options := &Options{}
	proxyManager := New(nil, context.Background(), options)
	go proxyManager.Run()
	proxyRaw := make(map[uint64]*Raw)
	raw1 := &Raw{Url: "http://5.161.105.73:39002", Error: make(chan error)}
	proxyRaw[123456] = raw1
	raw2 := &Raw{Url: "http://5.161.114.158:39000", Error: make(chan error)}
	proxyRaw[1234567] = raw2
	raw3 := &Raw{Url: "http://5.161.108.116:39000", Error: make(chan error)}
	proxyRaw[12345678] = raw3
	proxyManager.ProxyDetailChan() <- proxyRaw
	for _, raw := range proxyRaw {
		select {
		case err := <-raw.Error:
			if err != nil {
				fmt.Println(err)
				fmt.Println(raw.Url)
			}
		}
	}
	fmt.Println(proxyManager.proxies[123456])
}
