FROM golang:1.11 AS builder
ENV GOPATH=/go
WORKDIR /go/src/github.com/kubevirt/cluster-network-addons-operator/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cluster-network-addons-operator github.com/kubevirt/cluster-network-addons-operator/cmd/manager
RUN CGO_ENABLED=0 GOOS=linux go build -o /manifest-templator github.com/kubevirt/cluster-network-addons-operator/tools/manifest-templator

FROM centos:centos7
ENV ENTRYPOINT=/entrypoint \
    OPERATOR=/cluster-network-addons-operator \
    MANIFEST_TEMPLATOR=/manifest-templator \
    CSV_TEMPLATE=/cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in \
    USER_UID=1001 \
    USER_NAME=cluster-network-addons-operator
RUN \
    yum -y update \
    yum clean all
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/build/operator/bin/user_setup /user_setup
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/build/operator/bin/csv-generator /usr/bin/csv-generator
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/templates/cluster-network-addons/VERSION/cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in
RUN /user_setup
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/data /data
COPY --from=builder /cluster-network-addons-operator $OPERATOR
COPY --from=builder /manifest-templator $MANIFEST_TEMPLATOR
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/build/operator/bin/entrypoint $ENTRYPOINT
ENTRYPOINT $ENTRYPOINT
USER $USER_UID
