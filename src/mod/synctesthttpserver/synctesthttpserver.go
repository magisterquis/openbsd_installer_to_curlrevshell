// Package synctesthttpserver - Synctest-friendly mock HTTP server
package synctesthttpserver

/*
 * synctesthttpserver.go
 * Synctest-friendly mock HTTP server
 * By J. Stuart McMurray
 * Created 20260203
 * Last Modified 20260203
 */

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
)

/* Shamelessly mooched from https://go.dev/play/p/AVXzqqwiJPn */

type Server struct {
	*httptest.Server
	client *http.Client
}

func NewServer(h http.Handler) *Server {
	l := newListener()
	srv := &httptest.Server{
		Listener: l,
		Config:   &http.Server{Handler: h},
	}
	srv.Start()
	client := srv.Client()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		panic("httptest client transport is not *http.Transport")
	}
	transport.DialContext = l.DialContext
	return &Server{
		client: client,
		Server: srv,
	}
}

func (srv *Server) Client() *http.Client {
	return srv.client
}

type listener struct {
	addr   memAddr
	connCh chan net.Conn
	closed chan struct{}
	once   sync.Once
}

func newListener() *listener {
	return &listener{
		addr:   memAddr("test:80"),
		connCh: make(chan net.Conn),
		closed: make(chan struct{}),
	}
}

func (l *listener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *listener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *listener) Addr() net.Addr {
	return l.addr
}

func (l *listener) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case l.connCh <- server:
		return client, nil
	case <-l.closed:
		client.Close()
		server.Close()
		return nil, net.ErrClosed
	case <-ctx.Done():
		client.Close()
		server.Close()
		return nil, ctx.Err()
	}
}

type memAddr string

func (a memAddr) Network() string { return "mem" }
func (a memAddr) String() string  { return string(a) }
