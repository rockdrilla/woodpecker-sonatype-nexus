#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

r=0

TARGET_PLATFORMS=$(printf '%s' "${TARGET_PLATFORMS:?}" | tr ',' ' ')
for TARGET_PLATFORM in ${TARGET_PLATFORMS} ; do
    export TARGET_PLATFORM

    . .ci/envsh.build
    .ci/build.sh || r=$?
    [ "$r" = 0 ] || break
done

make ci-clean

exit "$r"
