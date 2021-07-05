# F5 IPAM Controller

The F5 IPAM Controller is a Docker container that runs in an orchestration environment and interfaces with an IPAM system.
It allocates IP addresses from an IPAM system’s address pool for hostnames in an orchestration environment.
The F5 IPAM Controller watches orchestration-specific resources and consumes the hostnames within each resource.

# In this IPAM

The F5 IPAM Controller can allocate IP address from static IP address pool based on the CIDR mentioned in a Kubernetes resource The idea here is that we will support CRD, Type LB and probably also in the future route/ingress. We should make it more generic so that we don't have to update this later, F5 IPAM Controller decides to allocate the IP from the respective IP address pool for the hostname specified in the virtualserver custom resource.

Supported kubernetes resource : 
| RESOURCES | MINIMUM VERSION SUPPORTED |
| ------ | ------ |
| VS CRD | CIS v2.2.2 | 



# Setup Diagram and Details

### Architectural diagram of how F5-IPAM-Controller(FIC) fits in the environment

![alt text](./image/img-1.png)
The F5 IPAM Controller acts as an interface to CIS to provide an IP address from a pool of IP's to each hostname provided in the virtual server CRD.

### Flow Chart for CIS-FIC working 
![alt text](./image/img-2.png)

### F5 IPAM Deploy Configuration Options

**Deployment Options**

| PARAMETER | TYPE | REQUIRED | DESCRIPTION |
| ------ | ------ | ------ | ------ |
| orchestration | String | Required | The orchestration parameter holds the orchestration environment i.e. Kubernetes. |
| ipam-provider | String | Required |  ipam-provider parameter holds the IP provider that holds the ownership of providing IP addresses such as infoblox, f5-ip-provider. Default is *f5-ip-provider*. |
| log-level | String | Optional |  Log level parameter specify various logging level such as DEBUG, INFO, WARNING, ERROR, CRITICAL. |

**Deployment Options of Provider (f5-ip-provider)**

| PARAMETER | TYPE | REQUIRED | DESCRIPTION |
| ------ | ------ | ------ | ------ |
| ip-range | String | Required |  ip-range parameter holds the IP address ranges and from this range, it creates a pool of IP address range which gets allocated to the corresponding hostname in the virtual server CRD |

**Deployment Options of Provider (infoblox)**

| PARAMETER | TYPE | REQUIRED | DESCRIPTION |
| ------ | ------ | ------ | ------ |
| infoblox-labels | String | Required | infoblox labels holds the mappings for infoblox's netView, dnsView and CIDR |
| infoblox-grid-host | String | Required |  URL (or IP Address) of Infoblox Grid Host |
| infoblox-wapi-port | String | Required | Port that the Infoblox Server listens on |
| infoblox-wapi-version | String | Required | Web API version of Infoblox
| infoblox-username | String | Required | Username of Infoblox User |
| infoblox-username | String | Required | Password of the given Infoblox User |


Note: On how to configure these Configuration Options, please refer to IPAM Deployment YAML example in below.

### Installation
#### RBAC -  ServiceAccount, ClusterRole and ClusterRoleBindings for F5 IPAM Controller

```
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ipam-ctlr-clusterrole
rules:
  - apiGroups: ["fic.f5.com"]
    resources: ["f5ipams","f5ipams/status"]
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

#### Example: F5 IPAM Controller Deployment YAML with Default Provider

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
        - --ip-range={"Dev":"172.16.3.21-172.16.3.30","Test":"172.  16.3.31-172.16.3.40", "Production":"172.16.3.41-172.16.3.50","Default":"172.16.3.51-172.16.3.60"}'
        - --log-level=DEBUG
        command:
        - /app/bin/f5-ipam-controller
        image: f5networks/f5-ipam-controller
        imagePullPolicy: IfNotPresent
        name: f5-ipam-controller
      serviceAccount: ipam-ctlr
      serviceAccountName: ipam-ctlr
```

#### Example: F5 IPAM Controller Deployment YAML with Infoblox Provider

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
        - --log-level=DEBUG
        - --ipam-provider
        - infoblox
        - --infoblox-labels
        - '{"Dev" :{"netView": "default", "dnsView": "default", "cidr": "172.16.4.0/24"},"Test" :{"netView": "test", "dnsView": "test", "cidr": "172.16.5.0/24"}}'
        - --infoblox-grid-host
        - 10.144.75.2
        - --infoblox-wapi-port=443
        - --infoblox-wapi-version
        - 2.11.2
        - --infoblox-username
        - user
        - --infoblox-password
        - paswd

        command:
        - /app/bin/f5-ipam-controller
        image: f5networks/f5-ipam-controller
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

To configure CIS to work with the F5 IPAM controller, the user needs to provide a parameter --ipam=true in the CIS deployment and also provide a parameter ipamLabel in the Kubernetes resource.

#### Note: ipamLabel can have values as mentioned in the ip-range parameter in the deployment.

#### Examples

**Virtual Server CR**

```
apiVersion: "cis.f5.com/v1"
kind: VirtualServer
metadata:
 name: coffee-virtual-server
 labels:
   f5cr: "true"
spec:
 host: coffee.example.com
 ipamLabel: Dev
 pools:
 - path: /coffee
   service: svc-2
   servicePort: 80
```


**Tansport Server CR**

```
  apiVersion: cis.f5.com/v1
  kind: TransportServer
  metadata:
    generation: 2
    labels:
      f5cr: "true"
  spec:
    ipamLabel: Test
    mode: standard
    pool:
      monitor:
        interval: 20
        timeout: 10
        type: tcp
      service: test-svc
      servicePort: 1344
    snat: auto
    type: tcp
    virtualServerPort: 1344
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


#### NOTE: 
- If the user provides the parameter --ipam=true in the CIS deployment, then CIS decides if it needs to retrieve an IP Address from the IPAM Controller or not 

- If a VirtualServer Address is specified in the Kubernetes resource, CIS will not leverage the IPAM Controller for IP address even if a ipamLabel parameter is specified.

- If No VirtualServer Address is specified in the Kubernetes resource and ipamLabel parameter is specified, CIS will leverage the IPAM Controller for allocation of IP address.


### Known Issues

- FIC does not allocate the last IP address specified in the ip     range.
- Updating the --ip-range in FIC deployment is an issue.
- Restarting FIC with infoblox ipam provider holds/allocate more ip addresses in infoblox.
