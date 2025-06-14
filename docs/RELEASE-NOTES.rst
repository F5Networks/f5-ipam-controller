Release Notes for F5 IPAM Controller for Kubernetes & OpenShift
=======================================================================

0.1.11
-------------

Added Functionality
```````````````````
**What’s new:**
    * Support for namespace to watch the multiple namespaces for IPAM CRD
    * Operator support for OpenShift 4.16

0.1.10
``````````````````````````

Vulnerability Fixes
```````````````````
CVE-2023-38545, CVE-2023-38546, CVE-2022-48337, CVE-2022-48338, CVE-2022-48339, CVE-2023-2491, CVE-2023-24329,
CVE-2023-40217, CVE-2023-4527, CVE-2023-4806, CVE-2023-4813, CVE-2023-4911, CVE-2023-44487, CVE-2023-28617,
CVE-2022-40897


Known Issues
`````````````
CVE-2024-2961

0.1.9
-------------
Added Functionality
```````````````````
**What’s new:**
    * Base image upgraded to RedHat UBI-9 for FIC Container image

Bug Fixes
````````````
* `Issue 2747 <https://github.com/F5Networks/k8s-bigip-ctlr/issues/2747>`_ Fix to persist IP addresses after CIS restart


0.1.8
-------------
Added Functionality
```````````````````
* Support for label with multiple IP ranges with comma seperated values :issues:`101`. See `documentation <https://raw.githubusercontent.com/F5Networks/f5-ipam-controller/main/docs/config_examples/f5-ip-provider/ipv4-addr-range-default-provider-deployment.yaml>`_

Bug Fixes
````````````
* :issues:`115` Reference handled properly in Database table

Known Issues
`````````````
* Appending new pool to existing range using the comma operator triggers FIC to reassign the newIP with new IP pool for the corresponding ipamLabel domains/keys

0.1.7
------------
Bug Fixes
`````````
* :issues:`98` IPAM Storage initialisation handled properly.

0.1.6
------------
Added Functionality
````````````
* Added support for
    - IPv6 address range support with default f5-ip-provider.


0.1.5
------------
Added Functionality
```````````````````
* F5 IPAM Controller supports InfoBlox(See `documentation <https://github.com/F5Networks/f5-ipam-controller/blob/main/README.md>`_)
* Persistent support added for F5 IPAM Controller default provider. So FIC now requires pvc with volume mounted in deployment for default provider(More details at `documentation <https://github.com/F5Networks/f5-ipam-controller/blob/main/README.md>`_).
* Added support for
    - Single NetView via deployment parameter `infoblox-netview`. It need not be provided via IPAM Label(See `documentation <https://github.com/F5Networks/f5-ipam-controller/blob/main/docs/config_examples/infoblox/infoblox-deployment.yaml>`_).
    - Standalone IP in Infoblox Provider.
    - `credentials-directory` configuration option for mounting infoblox credentials and self-signed certificate from kubernetes secrets.
* Disabled DNSView for Infoblox Provider(A - record support is deprecated)

Bug Fixes
`````````
* Stale status entries are cleared from IPAM custom resource.
* FIC restart allocates multiple IP addresses on InfoBlox

Known Issues
```````````
* With InfoBlox integration,
    * Update ip-range is not working as expected

Migration from 0.1.4
````````````````````
* `f5ipam` CRD is now renamed to `ipam`.
* Resource in clusterrole should be updated to ipam before upgrading to latest ipam(See latest clusterrole at `documentation <https://github.com/F5Networks/k8s-bigip-ctlr/blob/master/docs/config_examples/crd/Install/clusterrole.yml>`_)
* For F5 IPAM Controller default provider, update deployment with pvc and volume for persistance of DB.
  Volume mount is prerequisite for FIC v0.1.5(See `documentation <https://github.com/F5Networks/f5-ipam-controller/blob/main/README.md>`_ for FIC deploment with volume)



0.1.4
------------
Added Functionality
```````````````````
* F5 IPAM Controller supports InfoBlox (Preview - Available for VirtualServer CR only. See `documentation <https://github.com/F5Networks/f5-ipam-controller/blob/main/README.md>`_).

Known Issues
```````````
* With InfoBlox integration,
    * FIC restart allocates multiple IP addresses on InfoBlox
    * Update ip-range is not working as expected
    * TransportServer CR and Service Type LoadBalancer are not supported

0.1.3
-------------
Bug Fixes
`````````
* Old entries in IPAM CR spec/status are now removed when CIS gets restarted during VS update
* FIC does not allocate the last IP address specified in the ip range.
* Deleting resources releases IP address along with clearing corresponding spec entries.


0.1.2
-------------
Added Functionality
```````````````````
* FIC supports label-based IP address allocation.
* FIC is now compatible with k8s 1.20.
* FIC now creates the IPAM custom resource schema for validation.
* Earlier way of specifying --ip-range format is deprecated.

Known Issues
```````````
* FIC does not allocate the last IP address specified in the ip range.
* Updating the --ip-range in FIC deployment is an issue.

