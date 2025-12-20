#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024-2025, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

: "${IMAGE_NAME:?}"

[ -n "${EXTRA_TAGS}" ] || exit 0

. .ci/envsh.common
. .ci/envsh.registry

image_src="docker://${IMAGE_NAME}"

oci_dir="${PWD}/oci-layers"
image_interim="oci:${oci_dir}:$(basename "${IMAGE_NAME}"):${IMAGE_TAG}"

rm -rf "${oci_dir}" ; mkdir "${oci_dir}"

r=0

img_copy() {
    for i in $(seq 1 3) ; do
        skopeo copy --retry-times 2 --retry-delay 60s --all "$@" || { sleep 5 ; continue ; }
        return 0
    done
    return 1
}

while : ; do
    img_copy "${image_src}:${IMAGE_TAG}" "${image_interim}" || r=$?
    [ "$r" = 0 ] || break

    for tag in ${EXTRA_TAGS} ; do
        [ -n "${tag}" ] || continue

        echo " -> ${image_src}:${tag}"
        img_copy "${image_interim}" "${image_src}:${tag}" || r=$?
        [ "$r" = 0 ] || break
    done

    break
done

rm -rf "${oci_dir}"
exit "$r"
