#!/bin/sh
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

TARGET_PLATFORMS=$(printf '%s' "${TARGET_PLATFORMS:?}" | tr ',' ' ')
for TARGET_PLATFORM in ${TARGET_PLATFORMS} ; do
    r=0
    [ -n "${TARGET_PLATFORM}" ] || r=1
    [ "$r" = 0 ] || break
    export TARGET_PLATFORM

    .ci/env.sh || r=$?
    [ "$r" = 0 ] || break

    head -n 40 .build_env
    . ./.build_env
    [ -n "${GOARCH}" ] || r=1
    [ "$r" = 0 ] || break

    .ci/build.sh || r=$?
    [ "$r" = 0 ] || break
done

.ci/build-cleanup.sh

exit "$r"
