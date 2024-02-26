#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

## setup image registry authentication
export REGISTRY_AUTH_FILE="${PWD}/.auth.json"
[ -s "${REGISTRY_AUTH_FILE}" ] || exit 1

## produce _real_ BASE_IMAGE because "static-debian12:debug-nonroot" is not multiarch image (yet)
export BASE_IMAGE="${BASE_IMAGE:?}-${GOARCH:?}"

## build image
buildah bud \
    -t "${IMAGE_NAME}:${IMAGE_TAG}${PLATFORM_SUFFIX}" \
    -f ./Dockerfile.ci \
    ${IMAGE_MANIFEST:+ --manifest "${IMAGE_MANIFEST}" } \
    --platform "${TARGET_PLATFORM}" \
    --build-arg "TARGET_PLATFORM=${TARGET_PLATFORM}" \
    --build-arg "PLATFORM_SUFFIX=${PLATFORM_SUFFIX}" \
    --build-arg "BASE_IMAGE=${BASE_IMAGE}" \
    --build-arg "VERSION=${VERSION}" \
    --network=host
