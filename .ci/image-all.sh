#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

: "${TARGET_PLATFORMS:?}"

. .ci/envsh.common
. .ci/envsh.registry

: "${IMAGE_NAME:?}" "${IMAGE_TAG:?}"
IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"

## used by .ci/image.sh
export IMAGE_MANIFEST="${IMAGE}"

if buildah manifest exists "${IMAGE}" ; then
    buildah manifest rm "${IMAGE}"
fi
buildah manifest create "${IMAGE}"

r=0

TARGET_PLATFORMS=$(printf '%s' "${TARGET_PLATFORMS}" | tr ',' ' ')
for TARGET_PLATFORM in ${TARGET_PLATFORMS} ; do
    export TARGET_PLATFORM

    . .ci/envsh.build

    PLATFORM_SUFFIX='-'$(printf '%s' "${TARGET_PLATFORM}" | tr '/' '-')
    export PLATFORM_SUFFIX

    .ci/image.sh || r=$?
    [ "$r" = 0 ] || break

    buildah manifest add "${IMAGE}" "${IMAGE}${PLATFORM_SUFFIX}"
done

[ "$r" = 0 ] || exit "$r"

## list built image(s)
echo
echo 'IMAGES:'
echo
buildah images --all --noheading --format 'table {{.ID}} {{.Name}}:{{.Tag}} {{.Size}} {{.CreatedAtRaw}}' --filter "reference=${IMAGE_NAME}"
echo

buildah manifest push --all "${IMAGE}" "docker://${IMAGE}"
