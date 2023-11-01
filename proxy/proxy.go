package proxy

import (
	"bufio"
	"context"
	"errors"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

type Proxy struct {
	dialer proxy.Dialer
}

type httpProxyDialer struct {
	target  string
	forward proxy.Dialer
}

func HttpProxy(u *url.URL, forward proxy.Dialer) (proxy.Dialer, error) {
	d := &httpProxyDialer{
		target:  u.Host,
		forward: forward,
	}
	return d, nil
}

func (d *httpProxyDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *httpProxyDialer) DialContext(ctx context.Context, _, address string) (c net.Conn, err error) {
	if dialer, ok := d.forward.(proxy.ContextDialer); ok {
		c, err = dialer.DialContext(ctx, "tcp", d.target)
	} else {
		c, err = d.forward.Dial("tcp", d.target)
	}
	if err != nil {
		return nil, err
	}

	//p, _ := url.Parse(d.target)

	u, err := url.Parse("http://" + address)
	if err != nil {
		c.Close()
		return nil, err
	}

	u.Scheme = ""

	req, err := http.NewRequest(http.MethodConnect, u.String(), nil)
	if err != nil {
		c.Close()
		return nil, err
	}

	req.Close = false
	req.Header.Set("Proxy-Connection", "Keep-Alive")
	req.Header.Set("User-Agent", "arthas")

	if err := req.WithContext(ctx).Write(c); err != nil {
		c.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(c), req)
	if err != nil {
		resp.Body.Close()
		c.Close()
		return nil, err
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Close()
		return nil, errors.New(resp.Status)
	}

	return c, nil
}

func init() {
	proxy.RegisterDialerType("http", HttpProxy)
	proxy.RegisterDialerType("https", HttpProxy)
}
