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
COPY build/operator/bin/user_setup /user_setup
COPY build/operator/bin/csv-generator /usr/bin/csv-generator
COPY templates/cluster-network-addons/VERSION/cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in cluster-network-addons-operator.VERSION.clusterserviceversion.yaml.in
RUN /user_setup
COPY data /data
COPY build/_output/bin/manager $OPERATOR
COPY build/_output/bin/manifest-templator $MANIFEST_TEMPLATOR
COPY build/operator/bin/entrypoint $ENTRYPOINT
ENTRYPOINT $ENTRYPOINT
USER $USER_UID
