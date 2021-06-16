Release Notes for F5 IPAM Controller for Kubernetes & OpenShift
=======================================================================

Next Release
------------
Added Functionality
```````````````````
* Add New Provider Infoblox with IPV4 Support

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

