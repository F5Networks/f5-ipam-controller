# Frequently Asked Questions

This repository is an attempt to gather some frequently surfaced questions and provide some answers!

## General Questions

### What is FIC ?

The F5 IPAM Controller (FIC) is a container that runs in a orchestration environment. It allocates IP addresses from an IPAM system’s address pool for hostnames in an orchestration environment. The F5 IPAM Controller watches orchestration-specific resources and consumes the hostnames within each resource.

### Can CIS be deployed without using FIC ?

Yes. FIC can be used to automatically manage IP address allocation to CIS monitored resources. 

### Which CIS monitored resources are integrated with FIC ?

IngressLink, VirtualServer, TransportServer CRD, ServiceType LB. 

### Which IPAM providers are supported with FIC ? 

* f5-ip-provider 
  * Statically provide the pool of IP address range in the deployment based on an ipam label. Refer [examples](https://github.com/F5Networks/f5-ipam-controller/tree/main/docs/config_examples/f5-ip-provider)
* infoblox provider
  * Infoblox labels in deployment holds the mappings for Infoblox’s netView, dnsView, and CIDR. Refer [examples](https://github.com/F5Networks/f5-ipam-controller/tree/main/docs/config_examples/infoblox)

### Should IPAM CRD be created manually ?

If IPAM CRD is not present, it is created when CIS pod starts. During upgrades, if CRD needs to be updated, consider deleting the CRD for CIS to re-create with latest schema. 

### How to troubleshoot FIC pod logs ?

`kubectl logs deploy/<name-of-ipam-deployment> -n kube-system -f`

### Error - `Unable to Update IPAM: kube-system/***  Error: ipams.fic.f5.com "***" not found`

Verify cluster role permissions for ipam resource. 

`kubectl describe clusterrole <name> -n kube-system | grep ipam`

Delete existing IPAM CRD and re-start both CIS and IPAM pods

If you are upgrading from CIS 2.5.0, FIC 0.1.4, then make sure f5ipams CRD is unlinked completely.

### Error - `Unable to Establish Connection to DB, unable to open database file: no such file or directory`

Check for Persistent Volume user permissions.

### What to do when pod is stuck in `ContainerCreating` state for a long time?

It can most likely be a Persistent Volume issue. Check for events in FIC deployed `kube-system` namespace.

`kubectl get events -n kube system`

## IPAM PV Deployment

### When using Infoblox as Provider, do we still need to use persistentVolumes?

No. Volume mounts are needed to make default f5-ip-provider persistent (when using static IP range).


### Can I skip volumeMount even if I use default static f5-ip-provider?

Using volumeMount for persistent database is required starting with CIS v2.6.0 and FIC v1.5.0.


### What are Persistent DB storage requirements?

For 50 IPAddresses, file size is ~20KB.


### Can I use local storage volume for production environment?

Local persistent storage should only be considered for workloads that handle data replication and backup at the application layer. This makes the applications resilient to node or data failures and unavailability, despite the lack of such guarantees at the individual disk level.

Important limitations and caveats to consider when using Local Persistent Volumes:

Using local storage ties your application to a specific node, making your application harder to schedule. Applications which use local storage should specify a high priority so that lower priority pods, that don’t require local storage, can be preempted if necessary.
If that node or local volume encounters a failure and becomes inaccessible, then that pod also becomes inaccessible. Manual intervention, external controllers, or operators may be needed to recover from these situations.
While most remote storage systems implement synchronous replication, most local disk offerings do not provide data durability guarantees. This means that loss of the disk or node may result in loss of all the data on that disk.

### Independent of storage volume used, what is required for IPAM deployment?

Regardless of storage option used, IPAM controller expects a directory volume mount to `/app/ipamdb` path with read and write permission for IPAM controller user with UID 1200. This can be achieved using securityContext fsGroup or initContainers.


### How do I assign new IP addresses completely and remove old allocated IP addresses?

In the mount directory, rename or remove a file named `cis_ipam.sqlite3`.

## Upgrade notes

Any schema updates will be captured here.

| FIC version    | Description |
| ----------- | ----------- |
| from 0.1.5 to  >= 0.1.6    | IPv6 support is included with FIC. This needs an update in ipams CRD schema.        |
|  from 0.1.4 to  >= 0.1.5      | `f5ipam` CRD is renamed to `ipam`. Ensure deleting the older `f5ipam` CRD and any associated resources. Update clusterrole definition. <br/> If you are using static `f5-ip-provider `, volume mounts are needed for persistence. Refer examples for more details |

