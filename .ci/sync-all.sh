#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
# (c) 2024, Konstantin Demin
set -ef

[ -z "${CI_DEBUG}" ] || set -xv

: "${IMAGE_NAME:?}" "${EXT_IMAGE_NAME:?}"

. .ci/envsh.common
. .ci/envsh.registry

image_src="docker://${IMAGE_NAME}"
image_dst="docker://${EXT_IMAGE_NAME}"

oci_dir="${PWD}/oci-layers"
image_interim="oci:${oci_dir}:$(basename "${IMAGE_NAME}"):${IMAGE_TAG}"

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
    img_copy "${image_src}:${IMAGE_TAG}" "${image_interim}" || r=$?
    [ "$r" = 0 ] || break

    echo " -> ${image_dst}:${IMAGE_TAG}"
    img_copy "${image_interim}" "${image_dst}:${IMAGE_TAG}" || r=$?
    [ "$r" = 0 ] || break

    for tag in ${EXTRA_TAGS} ; do
        [ -n "${tag}" ] || continue

        echo " -> ${image_src}:${tag}"
        img_copy "${image_interim}" "${image_src}:${tag}" || r=$?
        [ "$r" = 0 ] || break

        echo " -> ${image_dst}:${tag}"
        img_copy "${image_interim}" "${image_dst}:${tag}" || r=$?
        [ "$r" = 0 ] || break
    done

    break
done

rm -rf "${oci_dir}"
exit "$r"
