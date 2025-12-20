ARG GO_IMAGE=docker.io/golang:1.25.5-trixie
ARG BASE_IMAGE=gcr.io/distroless/static-debian13:debug-nonroot

## ---

FROM ${GO_IMAGE} as build
SHELL [ "/bin/sh", "-ec" ]

ARG GOPROXY
ARG GOSUMDB
ARG GOPRIVATE

ARG RELMODE

WORKDIR /go/src

COPY . .

ENV GOMAXPROCS=4 \
    MALLOC_ARENA_MAX=4

RUN go env | grep -F -e GOPROXY -e GOSUMDB ; \
    make OUTDIR=/go/bin ; \
    make ci-clean

## ---

FROM ${BASE_IMAGE}

COPY --from=build /go/bin/publish-nexus /bin/

ENV GOMAXPROCS=4 \
    MALLOC_ARENA_MAX=4

ENTRYPOINT [ ]
CMD [ "/bin/publish-nexus" ]

USER nonroot:nonroot
