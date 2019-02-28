FROM centos:centos7

ENV OPERATOR=/usr/local/bin/cluster-network-addons-operator \
    USER_UID=1001 \
    USER_NAME=cluster-network-addons-operator

# install operator binary
COPY build/_output/bin/cluster-network-addons-operator ${OPERATOR}
COPY data /data

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
