# curlrevshell.mk
# Build things used for and with curlrevshell
# By J. Stuart McMurray
# Created 20260110
# Last Modified 20260205

CRS_CAFILE       = /etc/ssl/${CRS_CERT:T}
CRS_CERT        ?= ${TMPD}/crs_cert.pem
CRS_KEY         ?= ${TMPD}/crs_key.pem
CRS_TMPL         = crs.tmpl
CRS_TXTAR       ?= crs.txtar
OQA_BIN          = output_query_adapter
START_CALLBACKS  = ${TMPD}/start_callbacks.sh
START_SH         = start.sh

# Cert archive for curlrevshell.
${CRS_TXTAR}: ${CRS_CERT} ${CRS_KEY}
	echo "Generated $$(date)" >$@.tmp
	echo '-- cert --' >>$@.tmp
	cat ${CRS_CERT} >>$@.tmp
	echo '-- key --' >>$@.tmp
	cat ${CRS_KEY} >>$@.tmp
	mv $@.tmp $@

# TLS Certificate
${CRS_CERT}: ${CRS_KEY} ${CONFIG}
	openssl req\
		-new\
		-x509\
		-key ${>:M*.pem}\
		-days 3650\
		-nodes\
		-subj /CN=${TLS_CN}\
		-out $@.tmp
	mv $@.tmp $@
# Mooched from https://tales.mbivert.com/on-letsencrypt-on-openbsd/

# TLS Key
${CRS_KEY}:
	openssl genrsa -out $@.tmp 4096
	mv $@.tmp $@

.poison empty (CRS_CBADDR)

# Launcher and template
${TMPD}/${START_SH} ${START_CALLBACKS} ${CRS_TMPL}: src/${@F}.m4 ${CONFIG}
	m4\
		-PEE\
		-Dm4_cafile=${CRS_CAFILE}\
		-Dm4_crs_cbaddr=${CRS_CBADDR}\
		-Dm4_crs_tmpl=${CRS_TMPL}\
		-Dm4_oqa_cbaddr=${OQA_CBADDR}\
		-Dm4_tls_txtar=${CRS_TXTAR}\
		${>:N*.mk} >$@.tmp
	mv $@.tmp $@

# Launcher needs the execute bit, though.
${START_SH}: ${TMPD}/${START_SH}
	cp $> $@
	chmod +x $@

# User-agent adapter
${OQA_BIN}: src/cmd/$@/$@
	cp $> $@
src/cmd/${OQA_BIN}/${OQA_BIN}!
	${.MAKE} -C ${@D} ${@F}
