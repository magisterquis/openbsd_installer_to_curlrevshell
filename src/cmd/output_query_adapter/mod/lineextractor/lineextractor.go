// Package lineextractor - Extract output lines from HTTP requests
package lineextractor

/*
 * lineextractor.go
 * Extract output lines from HTTP requests
 * By J. Stuart McMurray
 * Created 20260119
 * Last Modified 20260119
 */

import (
	"net/http"
	"net/url"
)

// ExtractLine extracts an output line from an HTTP request.
// It URL-decodes and returns the raw query query string, i.e. the part of
// the URL after the ?.
func ExtractLine(r *http.Request) (string, error) {
	return url.QueryUnescape(r.URL.RawQuery)
}
