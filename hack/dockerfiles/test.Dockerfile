ARG RUNC_VERSION=69663f0bd4b60df09991c08812a60108003fa340
ARG CONTAINERD_VERSION=430969e9fac6c57eb115a9fcb6c2e92ea44344c0
# containerd v1.0 for integration tests
ARG CONTAINERD10_VERSION=v1.0.3
# available targets: buildkitd, buildkitd.oci_only, buildkitd.containerd_only
ARG BUILDKIT_TARGET=buildkitd
ARG REGISTRY_VERSION=2.6

# The `buildkitd` stage and the `buildctl` stage are placed here
# so that they can be built quickly with legacy DAG-unaware `docker build --target=...`

FROM golang:1.10-alpine AS gobuild-base
RUN apk add --no-cache g++ linux-headers
RUN apk add --no-cache git libseccomp-dev make

FROM gobuild-base AS buildkit-base
WORKDIR /go/src/github.com/moby/buildkit
COPY . .

FROM buildkit-base AS buildctl
ENV CGO_ENABLED=0
ARG GOOS=linux
RUN go build -ldflags '-d' -o /usr/bin/buildctl ./cmd/buildctl

FROM buildkit-base AS buildkitd
ENV CGO_ENABLED=1
RUN go build -installsuffix netgo -ldflags '-w -extldflags -static' -tags 'seccomp netgo cgo static_build' -o /usr/bin/buildkitd ./cmd/buildkitd

# test dependencies begin here
FROM gobuild-base AS runc
ARG RUNC_VERSION
ENV CGO_ENABLED=1
RUN git clone https://github.com/opencontainers/runc.git "$GOPATH/src/github.com/opencontainers/runc" \
	&& cd "$GOPATH/src/github.com/opencontainers/runc" \
	&& git checkout -q "$RUNC_VERSION" \
	&& go build -installsuffix netgo -ldflags '-w -extldflags -static' -tags 'seccomp netgo cgo static_build' -o /usr/bin/runc ./

FROM gobuild-base AS containerd-base
RUN apk add --no-cache btrfs-progs-dev
RUN git clone https://github.com/dmcgowan/containerd.git /go/src/github.com/containerd/containerd
WORKDIR /go/src/github.com/containerd/containerd

FROM containerd-base as containerd
ARG CONTAINERD_VERSION
RUN git checkout -q "$CONTAINERD_VERSION" \
  && make bin/containerd \
  && make bin/containerd-shim \
  && make bin/ctr

# containerd v1.0 for integration tests
FROM containerd-base as containerd10
ARG CONTAINERD10_VERSION
RUN git checkout -q "$CONTAINERD10_VERSION" \
  && make bin/containerd \
  && make bin/containerd-shim

FROM buildkit-base AS unit-tests
COPY --from=runc /usr/bin/runc /usr/bin/runc
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/containerd* /usr/bin/


FROM buildkit-base AS buildkitd.oci_only
ENV CGO_ENABLED=1
RUN go build -installsuffix netgo -ldflags '-w -extldflags -static' -tags 'no_containerd_worker seccomp netgo cgo static_build' -o /usr/bin/buildkitd.oci_only ./cmd/buildkitd

FROM buildkit-base AS buildkitd.containerd_only
ENV CGO_ENABLED=0
RUN go build -ldflags '-d'  -o /usr/bin/buildkitd.containerd_only -tags no_oci_worker ./cmd/buildkitd

FROM registry:$REGISTRY_VERSION AS registry

FROM unit-tests AS integration-tests
COPY --from=containerd10 /go/src/github.com/containerd/containerd/bin/containerd* /opt/containerd-1.0/bin/
COPY --from=buildctl /usr/bin/buildctl /usr/bin/
COPY --from=buildkitd /usr/bin/buildkitd /usr/bin
COPY --from=registry /bin/registry /usr/bin

FROM gobuild-base AS cross-windows
ENV GOOS=windows
WORKDIR /go/src/github.com/moby/buildkit
COPY . .

FROM cross-windows AS buildctl.exe
RUN go build -o /buildctl.exe ./cmd/buildctl

FROM cross-windows AS buildkitd.exe
ENV CGO_ENABLED=0
RUN go build -o /buildkitd.exe ./cmd/buildkitd

FROM alpine AS buildkit-export
RUN apk add --no-cache git
VOLUME /var/lib/buildkit

# Copy together all binaries for oci+containerd mode
FROM buildkit-export AS buildkit-buildkitd
COPY --from=runc /usr/bin/runc /usr/bin/
COPY --from=buildkitd /usr/bin/buildkitd /usr/bin/
COPY --from=buildctl /usr/bin/buildctl /usr/bin/
ENTRYPOINT ["buildkitd"]

# Copy together all binaries needed for oci worker mode
FROM buildkit-export AS buildkit-buildkitd.oci_only
COPY --from=buildkitd.oci_only /usr/bin/buildkitd.oci_only /usr/bin/
COPY --from=buildctl /usr/bin/buildctl /usr/bin/
ENTRYPOINT ["buildkitd.oci_only"]

# Copy together all binaries for containerd worker mode
FROM buildkit-export AS buildkit-buildkitd.containerd_only
COPY --from=runc /usr/bin/runc /usr/bin/
COPY --from=buildkitd.containerd_only /usr/bin/buildkitd.containerd_only /usr/bin/
COPY --from=buildctl /usr/bin/buildctl /usr/bin/
ENTRYPOINT ["buildkitd.containerd_only"]

FROM alpine AS containerd-runtime
COPY --from=runc /usr/bin/runc /usr/bin/
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/containerd* /usr/bin/
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/ctr /usr/bin/
VOLUME /var/lib/containerd
VOLUME /run/containerd
ENTRYPOINT ["containerd"]

FROM buildkit-${BUILDKIT_TARGET}


