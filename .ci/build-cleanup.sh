#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

## cleanup build
GOCACHE=$(go env GOCACHE)
GOMODCACHE=$(go env GOMODCACHE)
rm -rf "${GOCACHE:?}" "${GOMODCACHE:?}"
