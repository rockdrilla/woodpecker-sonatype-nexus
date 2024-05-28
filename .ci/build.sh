#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

mkdir -p dist
OUTDIR=dist
OUTSFX='-'$(printf '%s' "${TARGET_PLATFORM:?}" | tr '/' '-')

export OUTDIR OUTSFX

idle() {
    nice -n +40 \
    chrt -i 0 \
    ionice -c 3 \
    "$@"
}
idle make clean build || make clean build
