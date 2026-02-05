#!/bin/ksh
m4_dnl
m4_dnl start.sh.m4
m4_dnl Start curlrevshell and the adatpter
m4_dnl By J. Stuart McMurray
m4_dnl Created 20260118
m4_dnl Last Modified 20260203
m4_changecom(xxxx)# Generated m4_esyscmd(date)m4_changecom(#)m4_dnl

case ${1-} in
        crs|curlrevshell) set -x; go run \
                -trimpath \
                -ldflags "-w -s" \
                github.com/magisterquis/curlrevshell@latest \
                -callback-address m4_crs_cbaddr \
                -template m4_crs_tmpl \
                -tls-certificate-cache m4_tls_txtar ;;
        oqa|output_query_adapter) set -x; ./output_query_adapter \
                -curlrevshell https://m4_crs_cbaddr/o \
                -tls m4_tls_txtar ;;
        *) cat >&2 <<_eof
Usage: $(basename "$0") curlrevshell|output_query_adapter

Starts curlrevshell or output_query_adatpter with the same values as baked
into the miniroot image.
_eof
                exit 10 ;;
esac

# vim: ft=sh
