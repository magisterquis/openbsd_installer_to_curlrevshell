package main

/*
 * output_query_adapter.go
 * Tests for output_query_adapter.go
 * By J. Stuart McMurray
 * Created 20260118
 * Last Modified 20260119
 */

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/magisterquis/curlrevshell/lib/sstls"
	"github.com/magisterquis/openbsd_installer_to_curlrevshell/src/cmd/output_query_adapter/mod/testlogger"
)

// Does our HTTP client work as expected?
func TestNewHTTPClient_TLSWorks(t *testing.T) {
	/* Server with self-signed cert. */
	var (
		tl, lb = testlogger.New()
		sech   = make(chan error, 1)
		svr    = http.Server{ErrorLog: tl}
	)
	l, err := sstls.Listen("tcp", "127.0.0.1:0", "", 0, "")
	if nil != err {
		t.Fatalf("Error starting listener: %s", err)
	}
	go func() { sech <- svr.Serve(l) }()
	t.Cleanup(func() {
		if err := svr.Shutdown(context.Background()); nil != err {
			t.Errorf("Shutting down server: %s", err)
		}
		if err := <-sech; nil != err &&
			!errors.Is(err, http.ErrServerClosed) {
			t.Errorf("Server returned error: %s", err)
		}
		lb.TestEmpty(t)
	})

	/* Make a request.  Should get a 404, but that tells us TLS worked. */
	c, err := newHTTPClient(l.Fingerprint)
	if nil != err {
		t.Fatalf("Could not make HTTP client: %s", err)
	}
	res, err := c.Get(fmt.Sprintf("https://%s", l.Addr()))
	if nil != err {
		t.Fatalf("Error making HTTP request: %s", err)
	}
	defer res.Body.Close()
	if got, want := res.StatusCode, http.StatusNotFound; got != want {
		t.Errorf(
			"Unexpected respnose status code\n got: %d\nwant: %d",
			got,
			want,
		)
	}
}
