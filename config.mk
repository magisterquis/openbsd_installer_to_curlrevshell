# config.mk
# Configuration for the build
# By J. Stuart McMurray
# Created 20260128
# Last Modified 20260205

##############################################################################
# These are the user-settable parameters for building the miniroot image,    #
# TLS certificate, and other such things needed to get a shell.              #
#                                                                            #
# After editing this config, it is a good idea to run make clean && make     #
#                                                                            #
# All of the below may also be set at build-time by passing them to make(1). #
##############################################################################

# CRS_CBADDR is the address to which to call back to curlrevshell.
# It should have the same domain or IP address as OQA_CBADDR.
CRS_CBADDR ?= 10.0.0.10:4444

# OQA_CBADDR is the address to which to call back to output_query_adapter.
# It should have the same domain or IP address as CRS_CBADDR.
OQA_CBADDR ?= ${CRS_CBADDR:C,:[[:digit:]]+$,,}:5555

# TLS_CN is the common name to put in the generated TLS certificate.
# It should be the same domain or IP address as OQA_CBADDR and CRS_CBADDR,
# and by default is CBADDR's domain/IP.
TLS_CN ?= ${CRS_CBADDR:C,:[[:digit:]]+,,}

# Arch is the architecture for which we're building the miniroot image.
ARCH ?= ${MACHINE_ARCH}

# Mirror is the mirror from which to download the distributed miniroot image.
MIRROR ?= https://cdn.openbsd.org/pub/OpenBSD

# Version is the OpenBSD version from which to make a miniroot image.
VERSION !?= uname -r
