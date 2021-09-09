Refer to example deployments with default f5-ip-provider in this directory.

## Static IPAM deployments with persistent volume mounts

Kubernetes supports a wide variety of storage options. Refer [link](https://kubernetes.io/docs/concepts/storage/volumes) for more details.

###### _Note:_ Example in this repo is just for demo purpose and is not suitable for production environment. Read through the limitations with each of the storage options and choose as per your production need. Please refer [cloudodcs](https://clouddocs.f5.com/containers/latest/userguide/ipam/) for more details.

###### _Note:_  Local storage ties your application to a specific node as mentioned in nodeAffinity of PV yaml deployment.

_Pre-requisite:_ Ensure mount directory (In example, `localstorage-pv-pvc-example.yaml`, /tmp/cis_ipam) to be present on node.
