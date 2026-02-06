package main

/*
 * handlers.go
 * HTTP handlers
 * By J. Stuart McMurray
 * Created 20260117
 * Last Modified 20260205
 */

import (
	"log"
	"net/http"

	"github.com/magisterquis/openbsd_installer_to_curlrevshell/src/cmd/output_query_adapter/mod/lineextractor"
)

// idParam is used to extract an ID from a URL path.
const idParam = "ID"

// LineHandler handles lines.  See ConnManager for more details.
type LineHandler interface {
	CloseConn(urlPath string) error
	KeepAlive(urlPath string) error
	Send(urlPath, line string) (bool, error)
}

// NewMux returns a new [http.ServeMux] connected to cMgr.
func NewMux(cMgr LineHandler) *http.ServeMux {
	return newMux(handler{
		logf:   log.Printf,
		debugf: Debugf,
		cMgr:   cMgr,
	})
}

// newMux does what NewMux says it does, but with a handler, for testing.
func newMux(h handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /close/{"+idParam+"}", h.handleClose)
	mux.HandleFunc("GET /keepalive/{"+idParam+"}", h.handleKeepAlive)
	mux.HandleFunc("GET /line/{"+idParam+"}", h.handleLine)

	return mux
}

// handler passes data to our HTTP handlers.
type handler struct {
	logf   func(string, ...any) /* Test-settable. */
	debugf func(string, ...any) /* Test-settable. */
	cMgr   LineHandler          /* Really a ConnectionManager. */
}

// handleLine handles an inbound output line.
func (h handler) handleLine(w http.ResponseWriter, r *http.Request) {
	var (
		ra = r.RemoteAddr
		id = r.PathValue(idParam)
	)
	/* Extract the line. */
	line, err := lineextractor.ExtractLine(r)
	if nil != err {
		h.logf("[%s] Error extracting line for %s: %s", ra, id, err)
		ec := http.StatusBadRequest
		http.Error(w, http.StatusText(ec), ec)
		return
	}
	/* Send it to the connection manager. */
	opened, err := h.cMgr.Send(id, line)
	if nil != err {
		h.logf("[%s] Error sending %q to %s: %s", ra, line, id, err)
		ec := http.StatusInternalServerError
		http.Error(w, http.StatusText(ec), ec)
		return
	}
	if opened {
		h.logf("[%s] Opened new connection for %s", ra, id)
	}
	h.debugf("[%s] Sent %q to %s", ra, line, id)
}

// handleClose handles a request to close a connection.
func (h handler) handleClose(w http.ResponseWriter, r *http.Request) {
	var (
		ra = r.RemoteAddr
		id = r.PathValue(idParam)
	)
	if err := h.cMgr.CloseConn(id); nil != err {
		h.logf("[%s] Error closing connection for %s: %s", ra, id, err)
		ec := http.StatusInternalServerError
		http.Error(w, http.StatusText(ec), ec)
		return
	}
	h.logf("[%s] Closed connection for %s", r.RemoteAddr, id)
}

// handleKeepAlive handles a request to keep a connection alive.
func (h handler) handleKeepAlive(w http.ResponseWriter, r *http.Request) {
	var (
		ra = r.RemoteAddr
		id = r.PathValue(idParam)
	)
	if err := h.cMgr.KeepAlive(id); nil != err {
		h.logf("[%s] Error keeping %s alive: %s", ra, id, err)
		ec := http.StatusInternalServerError
		http.Error(w, http.StatusText(ec), ec)
		return
	}
	h.debugf("[%s] KeepAlive: %s", r.RemoteAddr, id)
}
