FROM golang:1.10 AS builder
ENV GOPATH=/go
WORKDIR /go/src/github.com/kubevirt/cluster-network-addons-operator/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cluster-network-addons-operator github.com/kubevirt/cluster-network-addons-operator/cmd/manager

FROM centos:centos7
ENV ENTRYPOINT=/entrypoint \
    OPERATOR=/cluster-network-addons-operator \
    USER_UID=1001 \
    USER_NAME=cluster-network-addons-operator
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/build/operator/bin/user_setup /user_setup
RUN /user_setup
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/data /data
COPY --from=builder /cluster-network-addons-operator $OPERATOR
COPY --from=builder /go/src/github.com/kubevirt/cluster-network-addons-operator/build/operator/bin/entrypoint $ENTRYPOINT
ENTRYPOINT $ENTRYPOINT
USER $USER_UID
