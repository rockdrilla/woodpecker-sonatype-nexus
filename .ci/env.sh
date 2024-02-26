#!/bin/sh
set -ef

## authentication in image registry
printf '%s' "${REGISTRY_AUTH}" > .auth.json

[ -z "${CI_DEBUG}" ] || set -xv

## flush build env
: > .build_env
echo "unset GOAMD64 GOARM GOPPC64 GOMIPS GOMIPS64" >> .build_env
echo "set -a" >> .build_env

## shifty-nifty shell goodies :)

## do same thing as GitLab does for CI_COMMIT_REF_SLUG:
## 1. lowercase string
## 2. replace not allowed chars with '-' (squeezing repeats)
##    allowed chars are: `0-9`, `a-z` and '-'
## 3. remove leading and trailing '-' (if any)
## 4. truncate string up to 63 chars
## 5. remove trailing '-' (if any)
ref_slug() {
    printf '%s' "${1:?}" \
    | sed -Ez 's/^(.+)$/\L\1/;s/[^0-9a-z]+/-/g;s/^-//;s/-$//;s/^(.{1,63}).*$/\1/;s/-$//' \
    | tr -d '\0'
}

## normalize image tag
## performs like ref_slug() except:
## - symbols '_' and '.' are allowed too
## - truncate string up to 96 chars
## - squeeze symbol sequences:
##   - '-' has higher priority than surrounding (leading and trailing) symbols
##   - first symbol in sequence has higher priority than following symbols
## NB: implementation is rather demonstrative than effective
image_tag_norm() {
    printf '%s' "${1:?}" \
    | sed -Ez 's/^(.+)$/\L\1/;s/[^0-9a-z_.]+/-/g' \
    | sed -Ez 's/\.+/./g;s/_+/_/g;s/[_.]+-/-/g;s/-[_.]+/-/g;s/([_.])[_.]+/\1/g' \
    | sed -Ez 's/^[_.-]//;s/[_.-]$//;s/^(.{1,95}).*$/\1/;s/[_.-]$//' \
    | tr -d '\0'
}

## produce GOOS and GOARCH from TARGET_PLATFORM
unset GOOS GOARCH _variant
IFS=/ read -r GOOS GOARCH _variant <<-EOF
${TARGET_PLATFORM}
EOF
## verify that GOOS and GOARCH are not empty
: "${GOOS:?}" "${GOARCH:?}"
## fill .build_env with Go-related variables
echo "GOOS=${GOOS}"     >> .build_env
echo "GOARCH=${GOARCH}" >> .build_env
if [ -n "${_variant}" ] ; then
    case "${GOARCH}" in
    amd64 )
        echo "GOAMD64=${_variant}" >> .build_env ;;
    arm )
        _variant=${_variant#v}
        echo "GOARM=${_variant}" >> .build_env
    ;;
    ppc64 | ppc64le )
        echo "GOPPC64=${_variant}" >> .build_env ;;
    mips | mipsle )
        echo "GOMIPS=${_variant}" >> .build_env ;;
    mips64 | mips64le )
        echo "GOMIPS64=${_variant}" >> .build_env ;;
    esac
fi

## misc CI things
# CI_COMMIT_SHORT_SHA="${CI_COMMIT_SHA:0:8}"
CI_COMMIT_SHORT_SHA=$(printf '%s' "${CI_COMMIT_SHA}" | cut -c 1-8)
echo "CI_COMMIT_SHORT_SHA=${CI_COMMIT_SHORT_SHA}" >> .build_env
CI_COMMIT_REF_SLUG="$(ref_slug "${CI_COMMIT_BRANCH}")"
if [ -n "${CI_COMMIT_SOURCE_BRANCH}" ] ; then
    CI_COMMIT_REF_SLUG="$(ref_slug "${CI_COMMIT_SOURCE_BRANCH}")"
fi
echo "CI_COMMIT_REF_SLUG=${CI_COMMIT_REF_SLUG}" >> .build_env

## image tag(s)
IMAGE_TAG="${CI_COMMIT_SHORT_SHA}-b${CI_PIPELINE_NUMBER}-${CI_COMMIT_REF_SLUG}"
if [ -n "${CI_COMMIT_SOURCE_BRANCH}" ] ; then
    echo "Running on branch '${CI_COMMIT_SOURCE_BRANCH}'"
else
    if [ "${CI_COMMIT_BRANCH}" != "${CI_REPO_DEFAULT_BRANCH}" ] ; then
        echo "Running on branch '${CI_COMMIT_BRANCH}'"
    else
        IMAGE_TAG="${CI_COMMIT_SHORT_SHA}"
    fi
fi
IMAGE_TAG=$(image_tag_norm "${IMAGE_TAG}")
echo "IMAGE_TAG=${IMAGE_TAG}" >> .build_env

## extra tag(s)
tags=$(image_tag_norm "branch-${CI_COMMIT_BRANCH}")
if [ -n "${CI_COMMIT_TAG}" ] ; then
    tags="${CI_COMMIT_TAG#v}"
fi
if [ "${CI_COMMIT_BRANCH}" = "${CI_REPO_DEFAULT_BRANCH}" ] ; then
    ## TODO: think about "latest" tag: it should be error-prone for "backward tag push"
    tags="${tags} ${VERSION} latest"
    echo "RELMODE=1" >> .build_env
fi
echo "EXTRA_TAGS='${tags}'" >> .build_env

echo "set +a" >> .build_env
