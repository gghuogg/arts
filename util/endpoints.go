package util

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Endpoint interface {
	// Host returns the IPv4/IPv6 address of a service.
	Host() string

	// Port returns the port of a service.
	Port() int

	// String formats and returns the Endpoint as a string.
	String() string
}

// LocalEndpoint implements interface Endpoint.
type LocalEndpoint struct {
	host string // host can be either IPv4 or IPv6 address.
	port int    // port is port as commonly known.
}

// Host returns the IPv4/IPv6 address of a service.
func (e *LocalEndpoint) Host() string {
	return e.host
}

// Port returns the port of a service.
func (e *LocalEndpoint) Port() int {
	return e.port
}

// String formats and returns the Endpoint as a string, like: 192.168.1.100:80.
func (e *LocalEndpoint) String() string {
	return fmt.Sprintf(`%s:%d`, e.host, e.port)
}

func NewEndpoint(address string) Endpoint {
	var array []string
	for _, v := range strings.Split(address, ":") {
		v = strings.Trim(v, "")
		if v != "" {
			array = append(array, v)
		}
	}
	if len(array) != 2 {
		panic(errors.New(`invalid address "%s" for creating endpoint, endpoint address is like "ip:port"`))
	}
	port, _ := strconv.Atoi(array[1])
	return &LocalEndpoint{
		host: array[0],
		port: port,
	}
}

func CalculateListenedEndpoints(address string) []Endpoint {
	var (
		endpoints = make([]Endpoint, 0)
	)
	var (
		addrArray     = strings.Split(address, ":")
		listenedIps   []string
		listenedPorts []int
	)
	if len(addrArray) == 1 {
		configItemName := "address"
		panic(errors.New(fmt.Sprintf(`invalid %s configuration "%s", missing port`,
			configItemName, address)))
	}
	// IPs.
	switch addrArray[0] {
	case "127.0.0.1":
		// Nothing to do.
	case "0.0.0.0", "":
		intranetIps, err := GetIntranetIpArray()
		if err != nil {
			log.Fatalf(`error retrieving intranet ip: %+v`, err)
			return nil
		}
		// If no intranet ips found, it uses all ips that can be retrieved,
		// it may include internet ip.
		if len(intranetIps) == 0 {
			allIps, err := GetIpArray()
			if err != nil {
				log.Fatalf(`error retrieving ip from current node: %+v`, err)
				return nil
			}
			listenedIps = allIps
			break
		}
		listenedIps = intranetIps
	default:
		listenedIps = []string{addrArray[0]}
	}
	// Ports.
	port, _ := strconv.Atoi(addrArray[1])
	listenedPorts = []int{port}
	for _, ip := range listenedIps {
		for _, port := range listenedPorts {
			endpoints = append(endpoints, NewEndpoint(fmt.Sprintf(`%s:%d`, ip, port)))
		}
	}
	return endpoints
}
