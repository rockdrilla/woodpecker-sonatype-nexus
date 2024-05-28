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

while : ; do
    skopeo copy --all "${image_src}:${IMAGE_TAG}" "${image_interim}" || r=$?
    [ "$r" = 0 ] || break

    skopeo copy --all "${image_interim}" "${image_dst}:${IMAGE_TAG}" || r=$?
    [ "$r" = 0 ] || break

    for tag in ${EXTRA_TAGS} ; do
        [ -n "${tag}" ] || continue

        skopeo copy --all "${image_interim}" "${image_src}:${tag}" || r=$?
        [ "$r" = 0 ] || break

        skopeo copy --all "${image_interim}" "${image_dst}:${tag}" || r=$?
        [ "$r" = 0 ] || break
    done

    break
done

rm -rf "${oci_dir}"
exit "$r"
