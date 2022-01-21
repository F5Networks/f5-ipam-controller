# Frequently Asked Questions

This guide is an attempt to gather some frequently surfaced questions and provide some answers!

## Table of Contents 
<!-- vscode-markdown-toc -->
* [General Questions](#GeneralQuestions)
	* [What is FIC ?](#WhatisFIC)
	* [Can CIS be deployed without using FIC ?](#CanCISbedeployedwithoutusingFIC)
	* [Which CIS monitored resources are integrated with FIC ?](#WhichCISmonitoredresourcesareintegratedwithFIC)
	* [Which IPAM providers are supported with FIC ?](#WhichIPAMprovidersaresupportedwithFIC)
	* [Should IPAM CRD be created manually ?](#ShouldIPAMCRDbecreatedmanually)
* [IPAM PV Deployment](#IPAMPVDeployment)
	* [When using Infoblox as Provider, do we still need to use persistentVolumes?](#WhenusingInfobloxasProviderdowestillneedtousepersistentVolumes)
	* [Can I skip volumeMount even if I use default static f5-ip-provider?](#CanIskipvolumeMountevenifIusedefaultstaticf5-ip-provider)
	* [What are Persistent DB storage requirements?](#WhatarePersistentDBstoragerequirements)
	* [Can I use local storage volume for production environment?](#CanIuselocalstoragevolumeforproductionenvironment)
	* [Independent of storage volume used, what is required for IPAM deployment?](#IndependentofstoragevolumeusedwhatisrequiredforIPAMdeployment)
	* [How do I assign new IP addresses completely and remove old allocated IP addresses?](#HowdoIassignnewIPaddressescompletelyandremoveoldallocatedIPaddresses)
* [Troubleshooting](#Troubleshooting)
	* [How to troubleshoot FIC pod logs ?](#HowtotroubleshootFICpodlogs)
	* [Error - `Unable to Update IPAM: kube-system/***  Error: ipams.fic.f5.com "***" not found`](#Error-UnabletoUpdateIPAM:kube-systemError:ipams.fic.f5.comnotfound)
	* [Error - `Unable to Establish Connection to DB, unable to open database file: no such file or directory`](#Error-UnabletoEstablishConnectiontoDBunabletoopendatabasefile:nosuchfileordirectory)
	* [What to do when pod is stuck in `ContainerCreating` state for a long time?](#WhattodowhenpodisstuckinContainerCreatingstateforalongtime)
* [Upgrade notes](#Upgradenotes)

<!-- vscode-markdown-toc-config
	numbering=false
	autoSave=true
	/vscode-markdown-toc-config -->
<!-- /vscode-markdown-toc -->


## <a name='GeneralQuestions'></a>General Questions

### <a name='WhatisFIC'></a>What is FIC ?

The F5 IPAM Controller (FIC) is a container that runs in a orchestration environment. It allocates IP addresses from an IPAM system’s address pool for hostnames in an orchestration environment. The F5 IPAM Controller watches orchestration-specific resources and consumes the hostnames within each resource.

### <a name='CanCISbedeployedwithoutusingFIC'></a>Can CIS be deployed without using FIC ?

Yes. FIC can be used to automatically manage IP address allocation to CIS monitored resources. 

### <a name='WhichCISmonitoredresourcesareintegratedwithFIC'></a>Which CIS monitored resources are integrated with FIC ?

IngressLink, VirtualServer, TransportServer CRD, ServiceType LB. 

### <a name='WhichIPAMprovidersaresupportedwithFIC'></a>Which IPAM providers are supported with FIC ?

* f5-ip-provider 
  * Statically provide the pool of IP address range in the deployment based on an ipam label. Refer [examples](https://github.com/F5Networks/f5-ipam-controller/tree/main/docs/config_examples/f5-ip-provider)
* infoblox provider
  * Infoblox labels in deployment holds the mappings for Infoblox’s netView, dnsView, and CIDR. Refer [examples](https://github.com/F5Networks/f5-ipam-controller/tree/main/docs/config_examples/infoblox)

### <a name='ShouldIPAMCRDbecreatedmanually'></a>Should IPAM CRD be created manually ?

If IPAM CRD is not present, it is created when CIS pod starts. During upgrades, if CRD needs to be updated, consider deleting the CRD for CIS to re-create with latest schema. 


## <a name='IPAMPVDeployment'></a>IPAM PV Deployment

### <a name='WhenusingInfobloxasProviderdowestillneedtousepersistentVolumes'></a>When using Infoblox as Provider, do we still need to use persistentVolumes?

No. Volume mounts are needed to make default f5-ip-provider persistent (when using static IP range).


### <a name='CanIskipvolumeMountevenifIusedefaultstaticf5-ip-provider'></a>Can I skip volumeMount even if I use default static f5-ip-provider?

Using volumeMount for persistent database is required starting with CIS v2.6.0 and FIC v1.5.0.


### <a name='WhatarePersistentDBstoragerequirements'></a>What are Persistent DB storage requirements?

For 50 IPAddresses, file size is ~20KB.


### <a name='CanIuselocalstoragevolumeforproductionenvironment'></a>Can I use local storage volume for production environment?

Local persistent storage should only be considered for workloads that handle data replication and backup at the application layer. This makes the applications resilient to node or data failures and unavailability, despite the lack of such guarantees at the individual disk level.

Important limitations and caveats to consider when using Local Persistent Volumes:

Using local storage ties your application to a specific node, making your application harder to schedule. Applications which use local storage should specify a high priority so that lower priority pods, that don’t require local storage, can be preempted if necessary.
If that node or local volume encounters a failure and becomes inaccessible, then that pod also becomes inaccessible. Manual intervention, external controllers, or operators may be needed to recover from these situations.
While most remote storage systems implement synchronous replication, most local disk offerings do not provide data durability guarantees. This means that loss of the disk or node may result in loss of all the data on that disk.

### <a name='IndependentofstoragevolumeusedwhatisrequiredforIPAMdeployment'></a>Independent of storage volume used, what is required for IPAM deployment?

Regardless of storage option used, IPAM controller expects a directory volume mount to `/app/ipamdb` path with read and write permission for IPAM controller user with UID 1200. This can be achieved using securityContext fsGroup or initContainers.


### <a name='HowdoIassignnewIPaddressescompletelyandremoveoldallocatedIPaddresses'></a>How do I assign new IP addresses completely and remove old allocated IP addresses?

In the mount directory, rename or remove a file named `cis_ipam.sqlite3`.

## <a name='Troubleshooting'></a>Troubleshooting

### <a name='HowtotroubleshootFICpodlogs'></a>How to troubleshoot FIC pod logs ?

* For detailed logs, FIC deployment can be configured with debug log level mode `--log-level=DEBUG`
  ```
  kubectl logs deploy/<name-of-ipam-deployment> -n kube-system -f
  ```
### <a name='Error-UnabletoUpdateIPAM:kube-systemError:ipams.fic.f5.comnotfound'></a>Error - `Unable to Update IPAM: kube-system/***  Error: ipams.fic.f5.com "***" not found`

* Verify below cluster role permissions for ipam resource.

  ```
  kubectl describe clusterrole <fic-clusterrole-name> -n kube-system | grep ipam
  ipams.fic.f5.com/status   []    []    [get list update watch create patch delete]
  ipams.fic.f5.com          []    []    [get list update watch create patch delete]
  ```
* Delete existing IPAM CRD and re-start both CIS and IPAM pods
  * Scale down CIS controller Deployment
  ```
  kubectl scale deploy/<cis-deployment-name> -n kube-system --replicas=0
  ```
  * Scale down FIC controller Deployment
  ```
  kubectl scale deploy/<fic-deployment-name> -n kube-system --replicas=0
  ```
  * Delete IPAM CRD Schema. By default, CIS deployment creates IPAM CRD schema, if it is not found.
  ```
  $ kubectl get crd | grep ipams
  
  NAME                          CREATED AT
  ...
  ipams.fic.f5.com            2021-12-16T09:24:58Z
  ...
  ```
  ```
  kubectl delete crd ipams.fic.f5.com
  ``` 
  * If you are upgrading from CIS 2.5.0, FIC 0.1.4, then make sure f5ipams CRD is unlinked completely.
  ```
  $ kubectl get crd | grep f5ipams
  
  NAME                          CREATED AT
  ...
  f5ipams.fic.f5.com            2021-12-16T09:24:58Z
  ...
  ```
  ```
  kubectl delete crd f5ipams.fic.f5.com 
  ```
  * Scale up CIS controller Deployment
  ```
  kubectl scale deploy/<cis-deployment-name> -n kube-system --replicas=1
  ```
  * Scale up FIC controller Deployment
  ```
  kubectl scale deploy/<fic-deployment-name> -n kube-system --replicas=1
  ```
  * There is no need to manually create IPAM CRD. CIS will create it automatically if not found.

### <a name='Error-UnabletoEstablishConnectiontoDBunabletoopendatabasefile:nosuchfileordirectory'></a>Error - `Unable to Establish Connection to DB, unable to open database file: no such file or directory`

* Check whether the persistent volume path (mentioned in the PV deployment), exists in the storage volume specified. 
  ```
  ls -l /path/to/file_mentioned_in_pv_deployment
  ```

### <a name='WhattodowhenpodisstuckinContainerCreatingstateforalongtime'></a>What to do when pod is stuck in `ContainerCreating` state for a long time?

* It can most likely be a Persistent Volume issue. Check for events in FIC deployed `kube-system` namespace.

  ```
  kubectl get events -n kube-system
  ```
  Note the messages of `Warning` and `Error` type events and act accordingly. 

## <a name='Upgradenotes'></a>Upgrade notes

Any schema updates will be captured here.

| FIC version    | Description |
| ----------- | ----------- |
| from 0.1.5 to  >= 0.1.6    | <li> IPv6 support is included with FIC. This needs an update to ipams CRD schema. <br> i) Delete existing IPAM CRD schema and CIS will automatically deploy latest IPAM CRD, if not found </li> |
|  from 0.1.4 to  >= 0.1.5      | <li> `f5ipam` CRD is renamed to `ipam`. Ensure deleting the older `f5ipam` CRD and any associated resources. Update clusterrole definition. <li> If you are using static `f5-ip-provider `, volume mounts are needed for persistence. Refer examples for more details |

