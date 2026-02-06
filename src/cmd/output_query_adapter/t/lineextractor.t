#!/bin/ksh
#
# generate_cert.t
# Can we make a TLS cert archive?
# By J. Stuart McMurray
# Created 20260118
# Last Modified 20260206

set -euo pipefail

. t/shmore.subr

tap_plan 24

# Start the extractor server
go run ../../mod/lineextractor/lineextractorserver |&
cleanup() {
        # Stop the server.
        exec 3>&p
        exec 3>&-
        # Make sure it stopped.
        read -pr
        tap_is "$REPLY" "Goodbye." "Server said goodbye" "$0" $LINENO
        wait
        tap_pass "Child processes exited"
        tap_done_testing
}
trap cleanup EXIT

# Get the listen address.
read -pr
tap_like \
        "$REPLY" \
        '^Listening on 127\.0\.0\.1:\d+$' \
        "Got listen address line" \
        "$0" $LINENO
ADDR=${REPLY#Listening on }
tap_like "$ADDR" '^127\.0\.0\.1:\d+$' "Address looks ok" "$0" $LINENO

# check checks if the server prints $1 after a request is with $1 as the query
# in the URL sent to the server with ftp(1).
# check emits two TAP lines.
#
# Arguments:
# $1 - Line to exfil
# $2 - Test name
# $3 - $LINENO
check() {
        local _line=$1 _name=$2 _lineno=$3
        set +e
        local _got=$(ftp -V -M -o - "http://$ADDR/?$_line")
        local _ret=$?
        set -e
        tap_is "$_ret"  0       "$_name - ftp(1) exited happily" "$0" "$_lineno"
        tap_is "$_got" "$_line" "$_name - Extraction correct"    "$0" "$_lineno"
}

check "kittens"   "Single word"                   $LINENO
check ""          "Empty string"                  $LINENO
check "?"         "Question mark"                 $LINENO
check "../../"    "Path traversal, leading dot"   $LINENO
check "/../../"   "Path traversal, leading slash" $LINENO
check "foo bar"   "Space in line"                 $LINENO
check " "         "Single space"                  $LINENO
check "# Comment" "Comment line"                  $LINENO
check "foo # bar" "Line with comment"             $LINENO
check "   foo"    "Leading spaces"                $LINENO


# vim: ft=sh
