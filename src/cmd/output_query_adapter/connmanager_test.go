package main

/*
 * connmanager_test.go
 * Manage persistent connections to the target
 * By J. Stuart McMurray
 * Created 20260117
 * Last Modified 20260206
 */

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/magisterquis/openbsd_installer_to_curlrevshell/src/mod/synctesthttpserver"
	"github.com/magisterquis/openbsd_installer_to_curlrevshell/src/mod/testlogger"
)

// Does the ConnManager work in the happy case?
func TestConnManager(t *testing.T) {
	/* Mock curlrevshell. */
	var (
		copyErr error
		gotID   string
		handled atomic.Bool
		nCopied int64
		oBuf    = new(bytes.Buffer)
		tl, lb  = testlogger.New()
		wg      sync.WaitGroup
	)
	wg.Add(1)
	mux := http.NewServeMux()
	mux.Handle("POST /o/{ID}", http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		/* Only do this once. */
		if !handled.CompareAndSwap(false, true) {
			t.Errorf("Handler called more than once")
		}
		defer wg.Done()

		/* Bidirectional comms. */
		rc := http.NewResponseController(w)
		if copyErr = rc.EnableFullDuplex(); nil != copyErr {
			return
		}
		if copyErr = rc.Flush(); nil != copyErr {
			return
		}
		/* Receive the output. */
		gotID = r.PathValue("ID")
		nCopied, copyErr = io.Copy(oBuf, r.Body)
	}))
	svr := httptest.NewServer(mux)
	t.Cleanup(func() { svr.CloseClientConnections(); svr.Close() })

	/* Can we send output? */
	var (
		cm           = NewConnManager(svr.URL+"/o", svr.Client())
		nCannedHaves = 10
		extraHaves   = []string{
			"        Significant whitespace",
			"",
		}
		haves      = make([]string, 0, nCannedHaves+len(extraHaves))
		wantOutput string
		id         = ts("id")
	)
	cm.logf = tl.Printf
	/* Make things to send. */
	for range nCannedHaves {
		n := len(haves) + 1
		line := fmt.Sprintf("Output %d - %s", n, ts("data"))
		haves = append(haves, fmt.Sprintf("%d\t%s", n, line))
		wantOutput += line + "\n"
	}
	/* Add a few more test case. */
	for _, line := range extraHaves {
		var t string
		if "" != line {
			t = "\t"
		}
		haves = append(
			haves,
			fmt.Sprintf("%d%s%s", len(haves)+1, t, line),
		)
		wantOutput += line + "\n"
	}

	/* Work out what we'll want. */
	for i, have := range haves {
		opened, err := cm.Send(id, have)
		if nil != err {
			t.Errorf(
				"Error sending line %d/%d: %s",
				i+1, len(haves),
				err,
			)
		}
		if 0 == i && !opened {
			t.Errorf("Connection not opened on first send")
		} else if 0 != i && opened {
			t.Errorf("Connection opened on send %d", i+1)
		}
	}

	/* Can we close the connection? */
	if err := cm.CloseConn(id); nil != err {
		t.Errorf("Error closing connection: %s", err)
	}

	/* Do we get an error if we close it again? */
	if err := cm.CloseConn(id); nil == err {
		t.Errorf("No error closing a closed connection")
	} else if !errors.Is(err, ErrNotOpen) {
		t.Errorf(
			"Incorrect error closing a closed connection: %s",
			err,
		)
	}

	/* Did it all work? */
	if handled.CompareAndSwap(false, true) {
		t.Fatalf("Handler not called")
	}
	wg.Wait()
	if nil != copyErr {
		t.Errorf("Error copying in handler: %s", copyErr)
	}
	if got, want := nCopied, int64(len(wantOutput)); got != want {
		t.Errorf(
			"Incorrect output size\n got: %d\nwant: %d",
			got,
			want,
		)
	}
	if got, want := gotID, id; got != want {
		t.Errorf(
			"Incorrect URL path handled\n got: %s\nwant: %s",
			got,
			want,
		)
	}
	if got, want := oBuf.String(), wantOutput; got != want {
		t.Errorf("Incorrect output\ngot:\n%q\nwant:\n%q", got, want)
	}
	lb.TestEmpty(t)
}

// Do keepalives work?
func TestConnManagerKeepAlive(t *testing.T) {
	synctest.Test(t, synctestConnManagerKeepAlive)
}

func synctestConnManagerKeepAlive(t *testing.T) {
	var (
		buf    bytes.Buffer
		hdone  = make(chan struct{})
		id     = ts("id")
		start  = time.Now()
		tl, lb = testlogger.New()
		svr    = synctesthttpserver.NewServer(http.HandlerFunc(func(
			_ http.ResponseWriter,
			r *http.Request,
		) {
			io.Copy(&buf, r.Body)
			close(hdone)
		}))
		cm    = NewConnManager(svr.URL, svr.Client())
		lines = []string{
			ts("open"),
			ts("still open"),
			ts("open after keepalive"),
		}
	)
	cm.logf = tl.Printf
	defer svr.Close()

	/* Start a connection. */
	if opened, err := cm.Send(id, "1 "+lines[0]); nil != err {
		t.Fatalf("Error sending open line: %s", err)
	} else if !opened {
		t.Errorf("Initial line did not open a connection")
	}
	defer cm.CloseConn(id)

	/* Wait a bit, should still be open. */
	time.Sleep(MaxKeepAliveWait - time.Nanosecond)
	synctest.Wait()
	if opened, err := cm.Send(id, "2 "+lines[1]); nil != err {
		t.Fatalf("Error sending still open line: %s", err)
	} else if opened {
		t.Errorf("Still open line opened a connection")
	}

	/* Send a keepalive, should keep it open. */
	if err := cm.KeepAlive(id); nil != err {
		t.Fatalf("Error sending keepalive: %s", err)
	}
	time.Sleep(MaxKeepAliveWait - time.Nanosecond)
	synctest.Wait()
	if opened, err := cm.Send(id, "3 "+lines[2]); nil != err {
		t.Fatalf("Error sending after keepalive line: %s", err)
	} else if opened {
		t.Errorf("Keepalive line opened a connection")
	}

	/* Let time out. */
	time.Sleep(2*time.Nanosecond + MaxKeepAliveWait)
	synctest.Wait()
	if opened, err := cm.Send(id, "4 should fail"); nil == err && opened {
		t.Errorf("Line number 4 started a new connection")
	} else if nil == err && !opened {
		t.Errorf("Connection did not time out")
	} else if !errors.Is(err, ErrNoConnection) {
		t.Fatalf(
			"Unexpected error sending after keepalive line: %s",
			err,
		)
	}

	/* We should have shut down by now. */
	<-hdone
	if got, want := time.Since(start), 3*MaxKeepAliveWait; got != want {
		t.Errorf(
			"Connection timed out at incorrect time\n"+
				" got: %s\n"+
				"want: %s",
			got,
			want,
		)
	}
	if got, want := buf.String(), strings.Join(
		lines,
		"\n",
	)+"\n"; got != want {
		t.Errorf(
			"Incorrect lines received\n got:\n%s\nwant:\n%s",
			got,
			want,
		)
	}

	/* Did logging work? */
	lb.TestStartsWith(t, fmt.Sprintf(
		"Closed connection for %s after timeout",
		id,
	))
	lb.TestEmpty(t)
}

// Do sent lines keep a connection alive?
func TestConnManagerSend_KeepAlive(t *testing.T) {
	synctest.Test(t, synctestConnManagerSendKeepAlive)
}

func synctestConnManagerSendKeepAlive(t *testing.T) {
	var (
		svr = synctesthttpserver.NewServer(http.HandlerFunc(func(
			_ http.ResponseWriter,
			r *http.Request,
		) {
			io.Copy(io.Discard, r.Body)
		}))
		cm     = NewConnManager(svr.URL, svr.Client())
		id     = ts("id")
		lineN  int
		start  = time.Now()
		tl, lb = testlogger.New()
	)
	cm.logf = tl.Printf
	defer svr.Close()

	/* sendLine sends the next numbered line. */
	sendLine := func() (bool, error) {
		lineN++
		return cm.Send(id, fmt.Sprintf("%d %s", lineN, ts("line")))
	}

	/* First line should make a connection. */
	if opened, err := sendLine(); nil != err {
		t.Fatalf("Error sending initial line: %s", err)
	} else if !opened {
		t.Errorf("Initial line did not open a connection")
	}
	defer cm.CloseConn(id)

	/* Send lines until the keepalive should expire. */
	for time.Now().Before(start.Add(MaxKeepAliveWait)) {
		time.Sleep(time.Second)
		synctest.Wait()
		if opened, err := sendLine(); nil != err {
			t.Fatalf("Error sending line %d: %s", lineN, err)
		} else if opened {
			t.Errorf("Line %d opened a connection", lineN)
		}
	}
	synctest.Wait() /* For just in case. */

	/* One last try, now that we should be expired. */
	if opened, err := sendLine(); nil != err {
		t.Fatalf("Error sending final line %d: %s", lineN, err)
	} else if opened {
		t.Errorf("Final line %d opened a connection", lineN)
	}

	/* Shouldn't have gotten any log lines. */
	lb.TestEmpty(t)
}

// ts returns s to which a hyped and a base36 uint64 have been appended.
func ts(s string) string {
	return fmt.Sprintf("%s-%s", s, strconv.FormatUint(rand.Uint64(), 36))
}
