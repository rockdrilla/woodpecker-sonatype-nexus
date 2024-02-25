ARG GO_IMAGE=docker.io/library/golang:1.21.7-bookworm
ARG BASE_IMAGE=gcr.io/distroless/static-debian12:debug-nonroot

## ---

FROM ${GO_IMAGE} as build
SHELL [ "/bin/sh", "-ec" ]

ARG GOPROXY
ARG GOSUMDB
ARG GOPRIVATE

ARG RELMODE

WORKDIR /go/src

COPY . .

ENV GOMAXPROCS=2 \
    MALLOC_ARENA_MAX=4

RUN go env | grep -F -e GOPROXY -e GOSUMDB -e GOPRIVATE ; \
    make OUTDIR=/go/bin ; \
    # cleanup intermediate layer
    eval "$(go env | grep -F -e GOCACHE -e GOMODCACHE)" ; \
    rm -rf ${GOCACHE} ${GOMODCACHE}

## ---

FROM ${BASE_IMAGE}

COPY --from=build /go/bin/plugin-sonatype-nexus /bin/

ENV GOMAXPROCS=4

CMD [ "/bin/plugin-sonatype-nexus" ]

USER nonroot:nonroot
