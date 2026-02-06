#!/bin/ksh
#
# generate_cert.t
# Can we make a TLS cert archive?
# By J. Stuart McMurray
# Created 20260118
# Last Modified 20260119

set -euo pipefail

. t/shmore.subr

tap_plan 2

TMPD=$(mktemp -d)
BIN="$TMPD/output_query_adapter"
TLS_ARCHIVE="$TMPD/crs.txtar"

# Star the server
go build -trimpath -ldflags "-w -s" -o "$TMPD/output_query_adapter" >/dev/null
$BIN \
        -debug \
        -listen "127.0.0.1:0" \
        -tls "$TLS_ARCHIVE" \
        2>&1 |&
SPID=$!

cleanup() {
        kill $SPID
        wait
        tap_pass "All children exited"
        rm -rf "$TMPD"
        tap_done_testing
}
trap cleanup EXIT

# Wait until we get the line that we've started.
read -pr
tap_like \
        "$REPLY" \
        'Serving HTTPS on 127.0.0.1:\d+$' \
        "Got Serving On line" \
        "$0" $LINENO

# vim: ft=sh
