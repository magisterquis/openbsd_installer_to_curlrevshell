// Program lineextractorserver - Test server for lineextractor
package main

/*
 * lineextractorserver.go
 * Test server for lineextractor
 * By J. Stuart McMurray
 * Created 20260119
 * Last Modified 20260205
 */

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/magisterquis/curlrevshell/lib/pledgeunveil"
	"github.com/magisterquis/openbsd_installer_to_curlrevshell/src/mod/lineextractor"
)

func main() {
	/* Command-line flags. */
	var (
		lAddr = flag.String(
			"listen",
			"127.0.0.1:0",
			"Listen `address`",
		)
		debug = flag.Bool(
			"debug",
			false,
			"Print each raw request URI and extracted line, "+
				"quoted",
		)
	)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: %s [options]

Test server for lineextractor.

Accepts HTTP requests and sends the lines extracted by lineextractor back to
the client.

Terminates when stdin is closed.

Options:
`,
			filepath.Base(os.Args[0]),
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	pledgeunveil.MustPledge("inet stdio")

	/* Start the listener here, so we can print the listen address. */
	l, err := net.Listen("tcp", *lAddr)
	if nil != err {
		log.Fatalf("Error listening on %s: %s", *lAddr, err)
	}
	defer l.Close()
	fmt.Printf("Listening on %s\n", l.Addr())

	pledgeunveil.MustPledge("inet stdio")

	/* End the program when stdin dies. */
	go func() {
		if _, err := io.Copy(io.Discard, os.Stdin); nil != err {
			log.Printf("Unexpected error reading stdin: %s", err)
		}
		fmt.Printf("Goodbye.\n")
		os.Exit(0)
	}()

	/* Print lines sent in HTTP requests. */
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		/* Extract the line. */
		line, err := lineextractor.ExtractLine(r)
		if nil != err {
			msg := fmt.Sprintf("Error: %s", err)
			http.Error(w, msg, http.StatusInternalServerError)
			if *debug {
				fmt.Printf("%s\n", msg)
			}
			return
		}
		if *debug {
			fmt.Printf("%q\n%q\n", r.RequestURI, line)
		}

		/* Send it back. */
		fmt.Fprintf(w, "%s", line)
	})
	if err := http.Serve(l, nil); nil != err {
		log.Fatalf("Fatal error: %s", err)
	}
}
