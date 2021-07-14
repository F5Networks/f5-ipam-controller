Release Notes for F5 IPAM Controller for Kubernetes & OpenShift
=======================================================================

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

