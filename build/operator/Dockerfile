ARG BUILD_ARCH=amd64
# Use official Go image instead of downloading Go manually
FROM --platform=linux/${BUILD_ARCH} docker.io/library/golang:1.25.7 AS builder

WORKDIR /go/src/cluster-network-addons-operator

# Copy go.mod and go.sum first to leverage Docker layer caching
# This allows dependency downloads to be cached separately from source code
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source code
COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o build/_output/bin/manager ./cmd/... && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o build/_output/bin/manifest-templator ./tools/manifest-templator/...

FROM --platform=linux/${TARGETARCH} quay.io/centos/centos:stream9 AS final

ENV ENTRYPOINT=/entrypoint \
    OPERATOR=/cluster-network-addons-operator \
    MANIFEST_TEMPLATOR=/manifest-templator \
    CSV_TEMPLATE=/cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in \
    USER_UID=1001 \
    USER_NAME=cluster-network-addons-operator \
    HOME=/home/${USER_NAME}

RUN \
    yum -y update && \
    yum clean all && \
    mkdir -p ${HOME} && \
    chown ${USER_UID}:0 ${HOME} && \
    chmod ug+rwx ${HOME}

COPY build/operator/bin/csv-generator /usr/bin/csv-generator
COPY templates/cluster-network-addons/VERSION/cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in
COPY data /data

COPY --from=builder /go/src/cluster-network-addons-operator/build/_output/bin/manager $OPERATOR
COPY --from=builder /go/src/cluster-network-addons-operator/build/_output/bin/manifest-templator $MANIFEST_TEMPLATOR
COPY build/operator/bin/entrypoint $ENTRYPOINT

ENTRYPOINT $ENTRYPOINT
USER $USER_UID
