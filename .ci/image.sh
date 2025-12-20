#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024-2025, Konstantin Demin
set -ef

. .ci/envsh.registry

[ -z "${CI_DEBUG}" ] || set -xv

## produce _real_ BASE_IMAGE because "static-debian12:debug-nonroot" is not multiarch image (yet)
export BASE_IMAGE="${BASE_IMAGE:?}-${GOARCH:?}"

img_pull() {
    for i in $(seq 1 3) ; do
        buildah pull --retry 2 --retry-delay 60s "$@" || { sleep 5 ; continue ; }
        return 0
    done
    return 1
}

img_build() {
    for i in $(seq 1 3) ; do
        buildah bud "$@" || { sleep 5 ; continue ; }
        return 0
    done
    return 1
}

img_pull \
    --platform "${TARGET_PLATFORM}" \
"${BASE_IMAGE}"

## build image
img_build \
    -t "${IMAGE_NAME}:${IMAGE_TAG}${PLATFORM_SUFFIX}" \
    -f ./Dockerfile.ci \
    ${IMAGE_MANIFEST:+ --manifest "${IMAGE_MANIFEST}" } \
    --platform "${TARGET_PLATFORM}" \
    --build-arg "TARGET_PLATFORM=${TARGET_PLATFORM}" \
    --build-arg "PLATFORM_SUFFIX=${PLATFORM_SUFFIX}" \
    --build-arg "BASE_IMAGE=${BASE_IMAGE}" \
    --network=host
