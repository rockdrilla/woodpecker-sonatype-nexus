#!/bin/sh
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

: "${TARGET_PLATFORMS:?}"

## semi-dry run
TARGET_PLATFORM=linux/amd64 \
.ci/env.sh
. ./.build_env

## setup image registry authentication
export REGISTRY_AUTH_FILE="${PWD}/.auth.json"
[ -s "${REGISTRY_AUTH_FILE}" ] || exit 1

buildah login "${IMAGE_REGISTRY:?}"

: "${IMAGE_NAME:?}" "${IMAGE_TAG:?}"
IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"
export IMAGE_MANIFEST="${IMAGE}"

if buildah manifest exists "${IMAGE}" ; then
    buildah manifest rm "${IMAGE}"
fi
buildah manifest create "${IMAGE}"

TARGET_PLATFORMS=$(printf '%s' "${TARGET_PLATFORMS}" | tr ',' ' ')
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

    # PLATFORM_SUFFIX='-'$(printf '%s' "${TARGET_PLATFORM}" | tr '/' '-')
    PLATFORM_SUFFIX="-${GOOS:?}-${GOARCH:?}"
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
buildah images --all --noheading --format 'table {{.ID}} {{.Name}}:{{.Tag}} {{.Size}} {{.CreatedAtRaw}}'
echo

## push image(s) and manifest(s)
buildah manifest push --all "${IMAGE}" "docker://${IMAGE}"

for tag in ${EXTRA_TAGS} ; do
    [ -n "${tag}" ] || continue
    if [ "${tag}" = "${IMAGE_TAG}" ] ; then continue ; fi

    buildah manifest push --all "${IMAGE}" "docker://${IMAGE_NAME}:${tag}"
done
