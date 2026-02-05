# Makefile
# Build ALL the things
# By J. Stuart McMurray
# Created 20260108
# Last Modified 20260110

TMPD   = tmp
CONFIG = config.mk

# Include our user-editable config.
.include "config.mk"

# Real builder.
BUILD_MK   = ${TMPD}/build.mk

# If we have the real builder, not much to do.
.if exists(${BUILD_MK})
.include "${BUILD_MK}"

# If not, build it and try again.
.else

.BEGIN: # Avoid infinite recursion.
.ifdef MADE_BUILD_MK
	@echo "Infinite Recursion detected."
	@exit 10
.endif

# Try again after we've build build.mk.
${.TARGETS}: ${BUILD_MK}
	MADE_BUILD_MK=true ${.MAKE} ${.TARGETS}
.endif

# Builder-builder
${BUILD_MK}: src/mk/mount.m4 src/mk/${@F}.m4 .NOTMAIN
.if !exists(${TMPD})
	+mkdir -p "${TMPD}"
.endif
	+m4 -PEE $> >$@.tmp
	+mv $@.tmp $@
