#!/bin/ksh
#
# output.t
# Does this thing work?
# By J. Stuart McMurray
# Created 20260118
# Last Modified 20260205

set -euo pipefail

. t/shmore.subr

tap_plan 58

# This all assumes OpenBSD.
if [[ "OpenBSD" != "$(uname -s)" ]]; then
        tap_plan 0 "Not using OpenBSD"
        exit 0
fi

TMPD=$(mktemp -d)
TLS_ARCHIVE="$TMPD/crs.txtar"
TLS_CAFILE="$TMPD/cert.pem"
SPID=
RPID=
OQA_LOG="$TMPD/oqa.log"
CRS_LOG="$TMPD/crs.log"
CRS_STARTED=false
ID="ID-$RANDOM"
cleanup() {
        # Kill child processes. */
        if $CRS_STARTED; then
                exec 3>&p
                exec 3>&-
        fi
        for p in "$SPID" "$RPID"; do
                if [ -n "$p" ]; then
                        kill "$p"
                fi
        done
        wait
        tap_pass "All child processes exited"
        # Clean up files
        rm -rf "$TMPD"
        # All done
        tap_done_testing
}
trap cleanup EXIT

# Make sure we have certs and such.
set +e
make -f ../../mk/curlrevshell.mk \
        CRS_CERT="$TLS_CAFILE" \
        CRS_KEY="$TMPD/crs_key.pem" \
        CRS_TXTAR="$TLS_ARCHIVE" \
        TLS_CN="127.0.0.1" \
        "$TLS_ARCHIVE" "$TLS_CAFILE" >"$TMPD/mk.out" 2>&1
RET=$?
set -e
tap_is "$RET" 0 "Make succeeded" "$0" $LINENO
if [[ 0 -ne "$RET" ]]; then
        tap_diag "Make's output:"
        tap_diag "$(<"$TMPD/mk.out")"
fi
[[ -f "$TLS_ARCHIVE" ]]
tap_pass "Made TLS archive"
[[ -f "$TLS_CAFILE" ]]
tap_pass "Made TLS CA file"

# Start curlrevshell and get useful info about it.
go run "github.com/magisterquis/curlrevshell@v0.0.1-beta.8" \
        -ctrl-i '' \
        -listen-address "127.0.0.1:0" \
        -log '' \
        -no-timestamps \
        -serve-files-from '' \
        -template '' \
        -tls-certificate-cache "$TLS_ARCHIVE" 2>"$TMPD/crs.err" |&
CRS_STARTED=true
read -pr
tap_is \
        "$REPLY" \
        "Welcome to curlrevshell version v0.0.1-beta.8" \
        "Got curlrevshell welcome message" \
        "$0" $LINENO
read -pr
tap_like \
        "$REPLY" \
        '^Listening on 127.0.0.1:\d+$' \
        "Got curlrevshell listening line" \
        "$0" $LINENO
ADDR=${REPLY#Listening on }
read -pr
tap_is "$REPLY" "To get a shell:" "Got To get a shell line" "$0" $LINENO
read -pr # Blank line
read -pr
tap_like "$REPLY" '^curl -sk ' "Got curl command" "$0" $LINENO
read -pr # Blank line
tap_is "$REPLY" "" "Got blank line waiting for a shell" "$0" $LINENO

# Write curlrevshell's output to a file for later examination.
>"$CRS_LOG"
( while read -pr; do echo "$REPLY" >>"$CRS_LOG"; done ) &
RPID=$!

# Start the server.
BIN="$TMPD/output_query_adapter"
go build -trimpath -ldflags "-w -s" -o "$TMPD/output_query_adapter" >/dev/null
$BIN \
        -curlrevshell "https://$ADDR/o" \
        -debug \
        -listen "127.0.0.1:0" \
        -tls "$TLS_ARCHIVE" \
        >"$OQA_LOG" 2>&1 &
SPID=$!

# Poll for the address.
while [[ 0 -eq $(($(wc -l <"$OQA_LOG"))) ]]; do
        sleep .1
done
read -r <"$OQA_LOG"
tap_like \
        "$REPLY" \
        'Serving HTTPS on 127.0.0.1:\d+$' \
        "Got output_query_adapter listen address" \
        "$0" $LINENO
ADDR=${REPLY#*Serving HTTPS on }

# send sends a line of output using OpenBSD's ftp(1).
# Each call to send emits two TAP lines.
#
# Arguments:
# $1 - Line name
# $2 - Line contents
send() {
        local _name=$1 _line=$2
        set +e
        GOT=$(ftp \
                -M \
                -o - \
                -S cafile="$TLS_ARCHIVE" \
                -V \
                "https://$ADDR/line/$ID?$_line" 2>&1)
        RET=$?
        set -e
        tap_is "$RET" "0" "$_name - ftp(1) exited happily" "$0" $LINENO
        tap_is "$GOT" ""  "$_name - ftp(1) exited silently" "$0" $LINENO
        return $RET
}
# Send a bunch of output
OUTPUT_LINE_COUNT=10
for i in `jot $OUTPUT_LINE_COUNT`; do
        LINE="Output line $i - ${RANDOM}_${RANDOM}"
        OUTPUT_LINES[$i]=$LINE
        send "Line $i" "$i $LINE"
done
# Plus ask to close the connection
set +e
GOT=$(ftp \
        -M \
        -o - \
        -S cafile="$TLS_ARCHIVE" \
        -V \
        "https://$ADDR/close/$ID?" 2>&1)
RET=$?
set -e
tap_is "$RET" 0  "FTP successfully requested connection close" "$0" $LINENO
tap_is "$GOT" "" "No output requesting connection close"       "$0" $LINENO

# line_is reads the next line from FD 3 and checks that it matches $1.
# line_is emits one TAP line.
#
# Arguments
# $1 - The want
# $2 - Test name
# $3 - Number of space-separated fields to remove at beginning of line
# $4 - $LINENO
line_is() {
        local _want=$1 _name=$2 _nfield=$3 _lineno=$4
        read -ru3 ||:
        local _got=$(echo -E "$REPLY" | cut -f "$((_nfield+1))"- -d ' ';)
        tap_is "$_got" "$_want" "$_name" "$0" "$_lineno"
}

# Check OQA's logs.
exec 3<"$OQA_LOG"
line_is \
        "Serving HTTPS on $ADDR" \
        "OQA log correct - Serving HTTPS on addres" \
        2 $LINENO
line_is \
        "Opened new connection for $ID" \
        "OQA log correct - Opened new connection" \
        3 $LINENO
for i in `jot "$OUTPUT_LINE_COUNT"`; do
        line_is \
                "Sent \"$i ${OUTPUT_LINES[$i]}\" to $ID" \
                "OQA log correct - Output line $i" \
                3 $LINENO
done
line_is \
        "Closed connection for $ID" \
        "OQA log correct - Closed connection" \
        3 $LINENO

# Check CRS's logs
exec 3<"$CRS_LOG"
line_is \
        "[127.0.0.1] Output connected: ID \"$ID\""\
        "CRS log correct - Output connected" \
        0 $LINENO
for i in `jot "$OUTPUT_LINE_COUNT"`; do
        line_is \
                "${OUTPUT_LINES[$i]}" \
                "CRS log correct - Output line $i" \
                0 $LINENO
done
line_is \
        "[127.0.0.1] Output connection closed" \
        "CRS log correct - Output connection closed" \
        0 $LINENO
line_is \
        "[127.0.0.1] Shell is gone :(" \
        "CRS log correct - Shell is gone" \
        0 $LINENO

# vim: ft=sh
