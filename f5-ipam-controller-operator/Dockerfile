# Build the manager binary
# FROM registry.redhat.io/openshift4/ose-helm-operator:v4.9
FROM registry.redhat.io/openshift4/ose-helm-operator:v4.11

ENV HOME=/opt/helm

### Required OpenShift Labels
LABEL name="F5 IPAM Controller Operator" \
      maintainer="f5_cis_operators@f5.com" \
      vendor="F5 Inc." \
      version="v0.0.2" \
      release="1" \
      summary="IPAM Controller Operator for F5 BIG-IP" \
      description="F5 IPAM Controller Operator is a Service Operator which installs F5 IPAM Controller on Kubernetes and OpenShift platforms and respective supported versions."


# Required Licenses
COPY f5-ipam-controller-operator/licenses /licenses

COPY f5-ipam-controller-operator/watches.yaml ${HOME}/watches.yaml
COPY helm-charts/f5-ipam-controller  ${HOME}/helm-charts/f5-ipam-controller
WORKDIR ${HOME}
