Release Notes for F5 IPAM Controller for Kubernetes & OpenShift
=======================================================================
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

