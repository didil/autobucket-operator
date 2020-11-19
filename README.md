# Autobucket Operator
Kubernetes Operator that automatically creates and manages Cloud Buckets (Object Storage) for k8s deployments. Built with Go and Operator SDK.

[![Build Status](https://travis-ci.org/didil/kubexcloud.svg?branch=master)](https://travis-ci.org/didil/kubexcloud)

**THIS SOFTWARE IS WORK IN PROGRESS / ALPHA RELEASE AND IS NOT MEANT FOR USAGE IN PRODUCTION SYSTEMS**

## Tests
To run tests:
````
$ make test
````

## Run locally
````
# deploy the k8s resources
$ make deploy
# run the operator locally
$ make run
````

## Usage
Deployment annotations sample:
````
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    ab.leclouddev.com/cloud: gcp
    ab.leclouddev.com/name-prefix: ab
    ab.leclouddev.com/on-delete-policy: destroy
````

- ````ab.leclouddev.com/cloud````: cloud where the storage bucket is created. Valid options: "gcp". If this annotation is missing or empty, no bucket is created for the deployment. 
- ````ab.leclouddev.com/name-prefix````: storage bucket name prefix. Default: "ab" (short name for autobucket). 
- ````ab.leclouddev.com/on-delete-policy````: bucket deletion policy when the deployment is deleted. Valid options: "ignore" (do nothing), "destroy" (delete the storage bucket). 
  
The full name format for the created storage buckets is "{prefix}-{namespace}-{deployment-name}"