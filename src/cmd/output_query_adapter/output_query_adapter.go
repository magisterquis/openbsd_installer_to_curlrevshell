// Program output_query_adapter - Adapter to convert ftp(1) HTTPS path query
// strings to curlrevshell input
package main

/*
 * output_query_adapter.go
 * Adapter to convert ftp(1) HTTPS path query strings to curlrevshell input
 * By J. Stuart McMurray
 * Created 20260111
 * Last Modified 20260124
 */

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/magisterquis/curlrevshell/lib/crsdialer"
	"github.com/magisterquis/curlrevshell/lib/pledgeunveil"
	"github.com/magisterquis/curlrevshell/lib/sstls"
)

// Debugf is log.Printf, but will be a no-op if -debug is not given.
var Debugf = log.Printf

func main() {
	/* Command-line flags. */
	var (
		lAddr = flag.String(
			"listen",
			"0.0.0.0:5555",
			"Listen `address`",
		)
		certFile = flag.String(
			"tls",
			"crs.txtar",
			"TLS certificate and key `archive`",
		)
		debugOn = flag.Bool(
			"debug",
			false,
			"Enable debug logging",
		)
		baseURL = flag.String(
			"curlrevshell",
			"https://127.0.0.1:4444/o",
			"Curlrevshell's base output `URL`",
		)
	)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: %s [options]

Adapter to convert ftp(1) HTTPS path query strings to curlrevshell input

Lines sent in query strings for a single connection  start with a number and
whitespace.  The first message on the connection should start with 1.  The
easiest way to do this is pass the output through cat -n.

The URL path should be
/line/{id}?line... for an output line
/close/{ID}        to close an output stream
/keepalive/{ID}    to keep a connection alive for another %s

Options:
`,
			filepath.Base(os.Args[0]),
			MaxKeepAliveWait,
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := pledgeunveil.Unveil(*certFile, "rwc"); nil != err {
		log.Fatalf("Error unveiling %s: %s", *certFile, err)
	}
	pledgeunveil.MustPledge("cpath inet rpath stdio wpath")

	/* Work out logging. */
	log.SetOutput(os.Stdout)
	if !*debugOn {
		Debugf = func(string, ...any) {}
	}

	/* Start TLS listener. */
	l, err := sstls.Listen("tcp", *lAddr, "", 0, *certFile)
	if nil != err {
		log.Fatalf("Error starting listener: %s", err)
	}

	pledgeunveil.MustPledge("inet stdio")

	/* Serve HTTP. */
	client, err := newHTTPClient(l.Fingerprint)
	if nil != err {
		log.Fatalf("Error setting up HTTP client: %s", err)
	}
	log.Printf("Serving HTTPS on %s", l.Addr())
	if err := http.Serve(
		l,
		NewMux(NewConnManager(*baseURL, client)),
	); nil != err {
		log.Fatalf("Fatal error: %s", err)
	}
}

// newHTTPClient rolls an http.Client which checks if connected TLS servers'
// certificates match the given fingerprint.
func newHTTPClient(fp string) (*http.Client, error) {
	/* Fingerprint verifier. */
	vc, err := crsdialer.TLSFingerprintVerifier(fp)
	if nil != err {
		return nil, fmt.Errorf(
			"setting up TLS fingerprint verification: %w",
			err,
		)
	}

	/* Transport. */
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.ExpectContinueTimeout = 0
	t.ForceAttemptHTTP2 = true
	t.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		VerifyConnection:   vc,
	}

	/* Client. */
	return &http.Client{Transport: t}, nil
}
