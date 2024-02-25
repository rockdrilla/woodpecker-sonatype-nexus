#!/bin/sh
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

mkdir -p dist
export OUTDIR=dist
export OUTSFX="-${GOOS:?}-${GOARCH:?}"

idle() {
    nice -n +40 \
    chrt -i 0 \
    ionice -c 3 \
    "$@"
}
idle make clean build || make clean build
