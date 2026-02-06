package main

/*
 * connmanager.go
 * Manage persistent connections to the target
 * By J. Stuart McMurray
 * Created 20260117
 * Last Modified 20260205
 */

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MaxKeepAliveWait is how long we wait to get a keepalive before closing a
// connection.
// It is three keepalives plus a second.
const MaxKeepAliveWait = 16 * time.Second

// lineRE matches the sort of line we expect in a user-agent string.
var lineRE = regexp.MustCompile(`^\s*(\d+)(?:\s(.*))?$`)

// ErrNotOpen is returned by ConnManager.Close when the connection to be closed
// wasn't actually open.
var ErrNotOpen = errors.New("connection not open")

// ErrNoConnection is returned when a non-first line was sent to an ID without
// an open connection.
var ErrNoConnection = errors.New("no exsiting connection")

type conn struct {
	pw  *io.PipeWriter
	kat *time.Timer /* KeepAlive Timer. */
}

// ConnManager sends lines to curlrevshell.
type ConnManager struct {
	mu sync.Mutex

	logf    func(string, ...any) /* Test-settable. */
	baseURL string
	client  *http.Client
	conns   map[string]conn /* id -> HTTP Connection */
}

// NewConnManager returns a new ConnManager, ready for use.
func NewConnManager(baseURL string, client *http.Client) *ConnManager {
	return &ConnManager{
		logf:    log.Printf,
		baseURL: strings.TrimRight(baseURL, "/") + "/",
		client:  client,
		conns:   make(map[string]conn),
	}
}

// Send sends the line and a newline to curlrevshell.
// The returned boolean is true if this caused a connection open.
func (cm *ConnManager) Send(id, line string) (bool, error) {
	/* Make sure our line is formatted correctly, and grab the number for
	if we need to make a new connection. */
	ms := lineRE.FindStringSubmatch(line)
	if 3 != len(ms) {
		return false, errors.New("invalid line")
	}
	lineN, err := strconv.Atoi(ms[1])
	if nil != err {
		return false, fmt.Errorf("parsing line number %s: %w", ms[1], err)
	}
	line = ms[2]

	/* Get the connection for this path.  We could probably make a
	per-connection lock, but for something which won't likely be hugely
	parallel, premature optimization. */
	cm.mu.Lock()
	defer cm.mu.Unlock()
	c, ok := cm.conns[id]
	if !ok {
		/* Don't have a connection.  We'll make a new one if this is
		the first line in the series. */
		if 1 != lineN {
			return false, fmt.Errorf(
				"cannot send line with number %d: %w",
				lineN,
				ErrNoConnection,
			)
		}
		/* Try to make a new connection. */
		c = cm.newConnection(id)
		cm.conns[id] = c
	}

	/* Write the line to the connection, or at least try. */
	if _, err := io.WriteString(c.pw, line+"\n"); nil != err {
		if cerr := cm.closeConn(id); nil != cerr {
			cm.logf("Error closing conn for %s: %s", id, cerr)
		}
		return false, fmt.Errorf("sending line: %w", err)
	}

	/* Got a line, so likely alive. */
	if err := cm.keepAlive(id); nil != err {
		cm.logf("Error resetting keepalive timer for %s: %s", id, err)
	}

	return !ok, nil
}

// CloseConn closes the conn for the given URL path, if one exists.
func (cm *ConnManager) CloseConn(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.closeConn(id)
}

// closeConn does what CloseConn says it does, but requires the caller to hold
// cm's lock.
func (cm *ConnManager) closeConn(id string) error {
	c, ok := cm.conns[id]
	if !ok {
		return ErrNotOpen
	}
	delete(cm.conns, id)
	return c.pw.Close()
}

// newConnection opens a new connection to the server.
func (cm *ConnManager) newConnection(id string) conn {
	/* Connect to the server. */
	pr, pw := io.Pipe()

	/* Proxy lines from the pipe. */
	go func() {
		var err error
		defer func() {
			pr.CloseWithError(err)
			pw.CloseWithError(err)
		}()
		res, err := cm.client.Post(cm.baseURL+id, "", pr)
		if nil != err {
			err = fmt.Errorf("sending POST request: %w", err)
			return
		}
		defer res.Body.Close()

		/* Make sure we got the go-ahead. */
		if http.StatusOK != res.StatusCode {
			err = fmt.Errorf(
				"got non-OK response status %s",
				res.Status,
			)
			return
		}
		io.Copy(io.Discard, res.Body)
		/* Connection is done, close it. */
		if err := cm.CloseConn(id); errors.Is(err, ErrNotOpen) {
			return
		} else if nil != err {
			cm.logf(
				"Error closing ended connection for %s: %s",
				id,
				err,
			)
			return
		}
		cm.logf("Connection for %s ended", id)
	}()

	/* Shut down the connection if nothing's kept it alive. */
	t := time.AfterFunc(MaxKeepAliveWait, func() {
		err := cm.CloseConn(id)
		if errors.Is(err, ErrNotOpen) {
			/* Normal. */
			return
		} else if nil != err {
			cm.logf(
				"Error closing connection for %s after "+
					"timeout: %s",
				id,
				err,
			)
			return
		}
		cm.logf("Closed connection for %s after timeout", id)
	})

	/* All looks good. */
	return conn{pw: pw, kat: t}
}

// KeepAlive resets id's keepalive timer, if it exists.
func (cm *ConnManager) KeepAlive(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.keepAlive(id)
}

// keepAlive does what keepAlive says it does, but requires the caller to hold
// cm's lock.
func (cm *ConnManager) keepAlive(id string) error {
	c, ok := cm.conns[id]
	if !ok {
		return ErrNotOpen
	}
	c.kat.Reset(MaxKeepAliveWait)
	return nil
}
