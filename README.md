# F5 IPAM Controller [ PREVIEW ]

The F5 IPAM Controller is a Docker container that runs in an orchestration environment and interfaces with an IPAM system.
It allocates IP addresses from an IPAM system’s address pool for hostnames in an orchestration environment.
The F5 IPAM Controller watches orchestration-specific resources and consumes the hostnames within each resource.

# In this Preview

The F5 IPAM Controller can allocate IP address from static IP address pool based on the CIDR mentioned in a Kubernetes resource The idea here is that we will support CRD, Type LB and probably also in the future route/ingress. We should make it more generic so that we don't have to update this later, F5 IPAM Controller decides to allocate the IP from the respective IP address pool for the hostname specified in the virtualserver custom resource.


### F5 IPAM Deploy Configuration Options

**Deployment Options**

| PARAMETER | TYPE | REQUIRED | DESCRIPTION |
| ------ | ------ | ------ | ------ |
| orchestration | String | Required | The orchestration parameter holds the orchestration environment i.e. Kubernetes. |
| ip-range | String | Required |  ip-range parameter holds the IP address ranges and from this range, it creates a pool of IP address range which gets allocated to the corresponding hostname in the virtual server CRD |
| log-level | String | Optional |  Log level parameter specify various logging level such as DEBUG, INFO, WARNING, ERROR, CRITICAL. |

***Example***

 ```
 - --orchestration=kubernetes
 - --ip-range=" 10.10.10.0/24-10.10.10.0/26"
 - --log-level=DEBUG // log-level can be INFO, DEBUG, WARNING, ERROR and CRITICAL.
```

#### RBAC -  ServiceAccount, ClusterRole and ClusterRoleBindings for F5 IPAM Controller
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

#### Example: F5 IPAM Controller Deployment YAML

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
        - --ip-range="10.10.10.0/24-10.10.10.0/26"
        - --log-level=DEBUG
        command:
        - /app/bin/f5-ipam-controller
        image: f5Networks/f5-ipam-controller
        imagePullPolicy: IfNotPresent
        name: f5-ipam-controller
      serviceAccount: ipam-ctlr
      serviceAccountName: ipam-ctlr
```

#### Deploying RBAC and F5 IPAM Controller 

Using kubectl let's apply the above defined RBAC and Deployment definitions.

```
kubectl create -f f5-ipam-rbac.yaml
kubectl create -f f5-ipam-deployment.yaml
```


### Configuring CIS to work with F5 IPAM Controller

To configure CIS to work with the F5 IPAM controller, the user needs to give a parameter ```--ipam=true``` in the CIS deployment and also provide a parameter ```cidr``` in the virtual server CRD.

#### Examples

**Virtual Server CRD**

```
apiVersion: "cis.f5.com/v1"
kind: VirtualServer
metadata:
 name: coffee-virtual-server
 labels:
   f5cr: "true"
spec:
 host: coffee.example.com
 cidr: "10.10.10.0/24"
 pools:
 - path: /coffee
   service: svc-2
   servicePort: 80
```

**CIS Deployment with ipam enabled**

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: k8s-bigip-ctlr-deployment
  namespace: kube-system
spec:
  replicas: 1
  template:
    metadata:
      name: k8s-bigip-ctlr
      labels:
        app: k8s-bigip-ctlr
    spec:
      serviceAccountName: bigip-ctlr
      containers:
        - name: k8s-bigip-ctlr
          image: "f5networks/k8s-bigip-ctlr"
          command: ["/app/bin/k8s-bigip-ctlr"]
          args: [
            "--bigip-username=$(BIGIP_USERNAME)",
            "--bigip-password=$(BIGIP_PASSWORD)",
            "--bigip-url=<ip_address-or-hostname>",
            "--bigip-partition=<name_of_partition>",
            "--pool-member-type=nodeport",
            "--agent=as3",
            "--ipam=true", //Enable IPAM 
            ]
      imagePullSecrets:
        - name: f5-docker-images
        - name: bigip-login
```


- NOTE: If the user provides the parameter ```--ipam=true``` in the CIS deployment then it is mandatory to provide the CIDR parameter in virtualserver CRD and also the virtualserver CRD should not have virtualServerAddress parameter.

### Updating the Status in Virtual Server CRD


The main aim of IPAM is to provide an IP address corresponding to each hostname provided in the VS CRD.

The user needs to mandatorily provide the host and CIDR in the hostSpecs section of F5-CR. The F5 IPAM Controller, in turn, reads the hostSpecs of CR, processes it, and updates the IPStatus with each host provided in the hostSpecs with host, IP(which is generated from the range of IP address by FIC), and corresponding CIDR.

- F5-ipam-controller (FIC) acts as a communication channel for updating the host, IP, and CIDR in VS CRD.

Below is the example: 

- f5-ipam-cr.yaml

```
apiVersion: "fic.f5.com/v1"
kind: F5IPAM
metadata:
  name: f5ipam.sample
  namespace: kube-system
spec:
  hostSpecs:
  - host: cafe.example.com
    cidr: 10.10.10.0/24
  - host: tea.example.com
    cidr: 10.10.10.0/24
status:
  IPStatus:
  - host: cafe.example.com
    ip: 10.10.10.45
    cidr: 10.10.10.0/24
  - host: tea.example.com
    ip: 10.10.10.47
    cidr: 10.10.10.0/24
```

 ### Limitations

- F5-ipam-controller cannot update and delete the hostname in the F5-IPAM custom resource hence update and deletion of IP address for virtual server custom may not work as expected. In case if the user wants to reflect the changes, the user can delete the F5-IPAM custom resource from kube-system named "f5ipam" and restart both the controller.
- Currently, F5 IPAM Controller does not support the update of CIDR and hostname.
- If F5-IPAM Controller is misconfigured after it allocates few IPs for VS CR. It will remove all its entry from the IPStatus. After the user reconfigured with the correct one, FIC may not get the previous same IPs for the hostname 



