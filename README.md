# CSI for S3
This is a Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md)) for S3 (or S3 compatible) storage. This can dynamically allocate buckets and mount them via a fuse mount into any container.

# Kubernetes installation
## Requirements
* Kubernetes 1.10+
* Kubernetes has to allow privileged containers
* Docker daemon must allow shared mounts (systemd flag `MountFlags=shared`)

## 1. Create a secret with your S3 credentials
The endpoint is optional if you are using something else than AWS S3. Also the region can be empty if you are using some other S3 compatible storage.
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: csi-s3-secret
stringData:
  accessKeyID: <YOUR_ACCESS_KEY_ID>
  secretAccessKey: <YOUR_SECRET_ACCES_KEY>
  endpoint: <S3_ENDPOINT_URL
  region: <S3_REGION>
```

## 2. Deploy the driver
```bash
cd deploy/kubernetes
$ kubectl create -f provisioner.yaml
$ kubectl create -f attacher.yaml
$ kubectl create -f csi-s3-driver.yaml
```

## 3. Create the storage class
```bash
$ kubectl create -f storageclass.yaml
```

## 4. Test the S3 driver
* Create a pvc using the new storage class:
```bash
$ kubectl create -f pvc.yaml
```
* Check if the PVC has been bound:
```bash
$ kubectl get pvc csi-s3-pvc
NAME         STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
csi-s3-pvc   Bound     pvc-c5d4634f-8507-11e8-9f33-0e243832354b   5Gi        RWX            csi-s3         9s
```
* Create a test pod which mounts your volume:
```bash
$ kubectl create -f poc.yaml
```
If the pod can start, everything should be working.

* Test the mount
```bash
$ kubectl exec -ti csi-s3-test-nginx bash
$ mount | grep fuse
s3fs on /var/lib/www/html type fuse.s3fs (rw,nosuid,nodev,relatime,user_id=0,group_id=0,allow_other)
$ touch /var/lib/www/html/hello_world
```
If something does not work as expected, check the troubleshooting section below.

# Additional configuration
## Mounter
By default the driver will use [s3fs](https://github.com/s3fs-fuse/s3fs-fuse) to mount buckets. Alternatively you can configure the storage class to use [goofys](https://github.com/kahing/goofys) for mounting S3 buckets. Note that goofys has some drawbacks in regards to POSIX compliance but in return offers better Performance than s3fs.

To configure a storage class to use goofys, just set the `mounter` parameter to `goofys`
```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: csi-s3
provisioner: ch.ctrox.csi.s3-driver
parameters:
  mounter: goofys
  csiProvisionerSecretName: csi-s3-secret
  csiProvisionerSecretNamespace: kube-system
```

# Limitations
As S3 is not a real file system there are some limitations to consider here. Depending on what mounter you are using, you will have different levels of POSIX compability. Also depending on what S3 storage backend you are using there are not always consistency guarantees. The detailed limitations can be found on the documentation of [s3fs](https://github.com/s3fs-fuse/s3fs-fuse#limitations) and [goofys](https://github.com/kahing/goofys#current-status).

# Troubleshooting
## Issues while creating PVC
* Check the logs of the provisioner:
```
$ kubectl logs -l app=csi-provisioner-s3 -c s3-csi-driver
```

## Issues creating containers
* Ensure feature gate `MountPropagation` is not set to `false`
* Check the logs of the s3-driver:
```
$ kubectl logs -l app=csi-s3-driver -c csi-s3-driver
```

# Development
This project can be built like any other go application.
```bash
$ go get -u github.com/ctrox/csi-s3-driver
```
## Build
```bash
$ make build
```
## Tests
Currently the driver is tested by the [CSI Sanity Tester](https://github.com/kubernetes-csi/csi-test/tree/master/pkg/sanity). As end-to-end tests require S3 storage and a mounter like s3fs, this is best done in a docker container. A Dockerfile and the test script are in the `test` directory. The easiest way to run the tests is to just use the make command:
```bash
$ make test
```
