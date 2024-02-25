#!/bin/sh
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

## cleanup build
GOCACHE=$(go env GOCACHE)
GOMODCACHE=$(go env GOMODCACHE)
rm -rf "${GOCACHE:?}" "${GOMODCACHE:?}"
