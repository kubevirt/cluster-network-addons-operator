ARG BUILD_ARCH=amd64
FROM --platform=linux/${BUILD_ARCH} quay.io/centos/centos:stream9 AS builder

RUN dnf install -y tar gzip jq && dnf clean all
RUN ARCH=$(uname -m | sed 's/x86_64/amd64/') && \
    GO_VERSION=$(curl -L -s "https://go.dev/dl/?mode=json" | jq -r '.[0].version') && \
    curl -L "https://go.dev/dl/${GO_VERSION}.linux-${ARCH}.tar.gz" -o go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
WORKDIR /go/src/cluster-network-addons-operator
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
