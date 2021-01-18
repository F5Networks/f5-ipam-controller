# f5-ipam-controller for CIS 2.2.2

The F5 IPAM Controller is a Docker container that runs in an orchestration environment and interfaces with an IPAM system. It allocates IP addresses from an IPAM system’s address pool for hostnames in an orchestration environment. The F5 IPAM Controller watches orchestration-specific resources and consumes the hostnames within each resource.

### The Controller can

Allocate IP address from static IP address pool based on the CIDR mentioned in a Kubernetes resource The idea here is that we will support CRD, Type LB and probably also in the future route/ingress. We should make it more generic so that we don't have to update this later, F5 IPAM Controller decides to allocate the IP from the respective IP address pool for the hostname specified in the virtualserver custom resource.


### F5 IPAM Deploy Configuration Options
 ```
 - --orchestration=kubernetes
 ```
The orchestration parameter holds the orchestration environment i.e. Kubernetes.
```
- --ip-range=" 172.16.3.17/28-172.16.3.30/28,172.16.3.33/28-172.16.3.46/28"
```
ip-range parameter holds the IP address ranges and from this range, it creates a pool of IP address range which gets allocated to the corresponding hostname in the virtual server CRD.
```
- --log-level=debug
```
Log level parameter specify various logging level such as DEBUG, INFO, WARNING, ERROR, CRITICAL.

#### Below is the RBAC for F5 IPAM Controller:
```
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ipam-ctlr-clusterrole
rules:
  - apiGroups: ["fic.f5.com"]
    resources: ["f5ipams"]
    verbs: ["get", "list", "watch", "update", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ipam-ctlr-clusterrole-binding
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ipam-ctlr-clusterrole
subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: ipam-ctlr
    namespace: kube-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ipam-ctlr
  namespace: kube-system
```

#### Deployment example:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    name: f5-ipam-controller
  name: f5-ipam-controller
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: f5-ipam-controller
  template:
    metadata:
      labels:
        app: f5-ipam-controller
    spec:
      containers:
      - args:
        - --orchestration=kubernetes
        - --ip-range="172.16.3.17/28-172.16.3.30/28,172.16.3.33/28-172.16.3.46/28"
        - --log-level=DEBUG
        command:
        - /app/bin/f5-ipam-controller
        image: f5Networks/f5-ipam-controller
        imagePullPolicy: IfNotPresent
        name: f5-ipam-controller
      serviceAccount: ipam-ctlr
      serviceAccountName: ipam-ctlr
```
#### Deploy RBAC and F5 IPAM Controller deployment
```
kubectl create -f f5-ipam-rbac.yaml
kubectl create -f f5-ipam-deployment.yaml
```


### Configuring CIS to work with F5 IPAM Controller


To configure CIS to work with the F5 IPAM controller, the user needs to give a parameter ```--ipam=true``` in the CIS deployment and also provide a parameter CIDR: "10.10.10.10/24" in the virtual server CRD.

- NOTE: If the user provides the parameter ```--ipam=true``` in the CIS deployment then it is mandatory to provide the CIDR parameter in virtualserver CRD and also the virtualserver CRD should not have virtualServerAddress parameter.

### Updating the Status in Virtual Server CRD


The main aim of IPAM is to provide an IP address corresponding to each hostname provided in the VS CRD.

The user needs to mandatorily provide the host and CIDR in the hostSpecs section of F5-CR. The F5 IPAM Controller, in turn, reads the hostSpecs of CR, processes it, and updates the IPStatus with each host provided in the hostSpecs with host, IP(which is generated from the range of IP address by FIC), and corresponding CIDR.

- F5-ipam-controller (FIC) acts as a communication channel for updating the host, IP, and CIDR in VS CRD.

 ### Limitations

- F5-ipam-controller cannot update and delete the hostname in the F5-IPAM custom resource hence update and deletion of IP address for virtual server custom may not work as expected. In case if the user wants to reflect the changes, the user can delete the F5-IPAM custom resource from kube-system named "f5ipam" and restart both the controller.
- Currently, F5 IPAM Controller does not support the update of CIDR and hostname.
- If F5-IPAM Controller is misconfigured after it allocates few IPs for VS CR. It will remove all its entry from the IPStatus. After the user reconfigured with the correct one, FIC may not get the previous same IPs for the hostname 



