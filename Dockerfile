# syntax=docker/dockerfile:1

ARG GO_VERSION=1.18
ARG GORELEASER_XX_BASE=crazymax/goreleaser-xx:edge
ARG XX_VERSION=master

FROM --platform=$BUILDPLATFORM ${GORELEASER_XX_BASE} AS goreleaser-xx
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-bullseye AS base
ENV CGO_ENABLED=1
COPY --from=goreleaser-xx / /
COPY --from=xx / /
RUN apt-get update \
    && apt-get install --no-install-recommends -y \
    clang \
    git \
    libtool \
    lld \
    pkg-config libc6-dev dpkg-dev

WORKDIR /src

FROM base AS build
ARG TARGETPLATFORM
RUN apt-get update
RUN xx-apt-get install -y \
    libgcc-10-dev libstdc++-10-dev libasound-dev portaudio19-dev libportaudio2 libportaudiocpp0 libaubio5 libaubio-dev 
# XX_CC_PREFER_STATIC_LINKER prefers ld to lld in ppc64le and 386.
ENV XX_CC_PREFER_STATIC_LINKER=1
RUN --mount=type=bind,source=.,rw \
    #    --mount=from=dockercore/golang-cross:xx-sdk-extras,target=/xx-sdk,src=/xx-sdk \
    # --mount=type=cache,target=/root/.cache \
    # goreleaser-xx --debug \
    # --go-binary="xx-go" \
    # --name="ledfx-$(xx-info debian-arch)" \
    # --dist="/out" \
    # --artifacts="archive" \
    # --artifacts="bin" \
    # --main="." \
    # --ldflags="-s -w " \
    # --envs="GO111MODULE=auto" \
    # --files="README.rst"
    xx-info env && xx-go build -v -o /out/ledfx-$(xx-info debian-arch)
FROM scratch AS artifact
COPY --from=build /out /

FROM scratch
COPY --from=build /usr/local/bin/ledfx /ledfx
ENTRYPOINT [ "/ledfx" ]