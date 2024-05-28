#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

: "${IMAGE_NAME:?}" "${EXT_IMAGE_NAME:?}" "${LATEST_TAG:?}"

. .ci/envsh.registry

image_src="docker://${IMAGE_NAME}"
image_dst="docker://${EXT_IMAGE_NAME}"

oci_dir="${PWD}/oci-layers"
image_interim="oci:${oci_dir}:$(basename "${IMAGE_NAME}"):${LATEST_TAG}"

rm -rf "${oci_dir}" ; mkdir "${oci_dir}"

r=0

img_copy() {
    for i in $(seq 1 3) ; do
        if skopeo copy --all "$@" ; then
            return 0
        fi
    done
    return 1
}

while : ; do
    img_copy "${image_src}:${LATEST_TAG}" "${image_interim}" || r=$?
    [ "$r" = 0 ] || break

    echo " -> ${image_src}:latest"
    img_copy "${image_interim}" "${image_src}:latest" || r=$?
    [ "$r" = 0 ] || break

    echo " -> ${image_dst}:latest"
    img_copy "${image_interim}" "${image_dst}:latest" || r=$?
    [ "$r" = 0 ] || break

    break
done

rm -rf "${oci_dir}"
exit "$r"
