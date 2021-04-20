Release Notes for F5 IPAM Controller for Kubernetes & OpenShift
=======================================================================

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

