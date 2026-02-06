package main

/*
 * handler_test.go
 * Tests for handler.go
 * By J. Stuart McMurray
 * Created 20260117
 * Last Modified 20260206
 */

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/magisterquis/openbsd_installer_to_curlrevshell/src/mod/testlogger"
)

// testLineHandler mocks ConnManager
type testLineHandler struct {
	mu     sync.Mutex
	closed bool
	open   bool
	buf    bytes.Buffer
}

func (lh *testLineHandler) CloseConn(_ string) error {
	lh.mu.Lock()
	defer lh.mu.Unlock()
	if lh.closed {
		return errors.New("already closed")
	}
	lh.closed = true
	lh.open = false
	return nil
}
func (lh *testLineHandler) Send(_, line string) (bool, error) {
	lh.mu.Lock()
	defer lh.mu.Unlock()
	if lh.closed {
		return false, errors.New("send after close")
	}
	lh.buf.WriteString(line + "\n")
	opened := !lh.open
	lh.open = true
	return opened, nil
}

func (lh *testLineHandler) KeepAlive(_ string) error {
	lh.mu.Lock()
	defer lh.mu.Unlock()
	if lh.closed {
		return errors.New("keepalive after close")
	}
	lh.buf.WriteString("keepalive\n")
	return nil
}

func TestHandler(t *testing.T) {
	var (
		tl, lb = testlogger.New()
		mgr    = new(testLineHandler)
		h      = handler{
			cMgr:   mgr,
			debugf: tl.Printf,
			logf:   tl.Printf,
		}
		mux   = newMux(h)
		haveN = 10 /* More haves. */
		haves = []string{
			"kittens",
			"# kittens",
			"/kittens?",
			"   kittens",
		}
		id      = ts("id")
		bufWant string
		logWant []string
	)

	/* Make a bunch of lines to send. */
	for i := range haveN {
		haves = append(
			haves,
			ts(fmt.Sprintf("Generated line %d", i+1)),
		)
	}

	/* Send a bunch of lines. */
	for i, have := range haves {
		/* String to send and sender. */
		bufWant += have + "\n"
		el := url.QueryEscape(have)
		req := httptest.NewRequest(
			http.MethodGet,
			"/line/"+id+"?"+el,
			nil,
		)
		/* First send will open the connection. */
		if 0 == i {
			logWant = append(logWant, fmt.Sprintf(
				"[%s] Opened new connection for %s",
				req.RemoteAddr,
				id,
			))
		}
		/* Update request to send the string. */
		logWant = append(logWant, fmt.Sprintf(
			"[%s] Sent %q to %s",
			req.RemoteAddr,
			have,
			id,
		))
		/* Send it forth. */
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		/* Should be a 200. */
		if got, want := rr.Code, http.StatusOK; got != want {
			t.Errorf(
				"Incorrect status sending line %d/%d\n"+
					" got: %d\n"+
					"want: %d",
				i+1, len(haves),
				got,
				want,
			)
		}
	}

	/* And a keepalive, to check logging. */
	req := httptest.NewRequest(http.MethodGet, "/keepalive/"+id, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	logWant = append(logWant, fmt.Sprintf(
		"[%s] KeepAlive: %s",
		req.RemoteAddr,
		id,
	))
	if got, want := rr.Code, http.StatusOK; got != want {
		t.Errorf(
			"Incorrect status sending keepalive\n"+
				" got: %d\n"+
				"want: %d",
			got,
			want,
		)
	}
	bufWant += "keepalive\n"

	/* Did it work? */
	mgr.mu.Lock()
	if got, want := mgr.buf.String(), bufWant; got != want {
		t.Errorf("Incorrect sent data\ngot:\n%s\nwant:\n%s", got, want)
	}
	if mgr.closed {
		t.Errorf("Manager closed before requesting close")
	}
	mgr.mu.Unlock()

	/* Can we close the connection. */
	req = httptest.NewRequest(http.MethodGet, "/close/"+id, nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	logWant = append(logWant, fmt.Sprintf(
		"[%s] Closed connection for %s",
		req.RemoteAddr,
		id,
	))
	if got, want := rr.Code, http.StatusOK; got != want {
		t.Errorf(
			"Incorrect status requesting close\n"+
				" got: %d\n"+
				"want: %d",
			got,
			want,
		)
	}
	mgr.mu.Lock()
	if !mgr.closed {
		t.Errorf("Manager not closed after requesting close")
	}
	mgr.mu.Unlock()

	/* Check logs. */
	lb.TestStartsWith(t, logWant...)
	lb.TestEmpty(t)
}
