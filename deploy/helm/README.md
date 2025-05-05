# Helm chart for csi-s3

This chart adds S3 volume support to your cluster.

## Install chart

```shell
helm install --namespace kube-system csi-s3 .
```

After installation succeeds, you can get a status of Chart: `helm status csi-s3`.

## Delete Chart

```shell
helm uninstall csi-s3 --namespace kube-system`
```

## Configuration

By default, this chart creates a secret and a storage class.

The following table lists all configuration parameters and their default values.

| Parameter                    | Description                                                            | Default                                                |
| ---------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------ |
| `storageClass.create`        | Specifies whether the storage class should be created                  | true                                                   |
| `storageClass.name`          | Storage class name                                                     | csi-s3                                                 |
| `storageClass.bucket`        | Existing bucket name to use, or leave blank to create                  |                                                        |
| `storageClass.usePrefix`     | Enable the prefix feature to avoid the removal of the prefix or bucket | false                                                  |
| `storageClass.prefix`        | can be empty (mounts bucket root), an existing prefix or a new one.    |                                                        |
| `storageClass.reclaimPolicy` | Volume reclaim policy                                                  | Delete                                                 |
| `storageClass.annotations`   | Annotations for the storage class                                      |                                                        |
| `secret.create`              | Specifies whether the secret should be created                         | true                                                   |
| `secret.name`                | Name of the secret                                                     | csi-s3-secret                                          |
| `secret.accessKey`           | S3 Access Key                                                          |                                                        |
| `secret.secretKey`           | S3 Secret Key                                                          |                                                        |
| `secret.endpoint`            | Endpoint                                                               | https://storage.yandexcloud.net                        |
