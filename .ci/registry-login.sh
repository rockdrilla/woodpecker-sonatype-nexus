#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

unset _bin
for i in podman buildah skopeo ; do
    if command -V "$i" >/dev/null ; then
        _bin=$i
        break
    fi
done
: "${_bin:?}"

. .ci/envsh.registry

for i ; do
    "${_bin}" login "$i" </dev/null
done
