#!/usr/bin/env bash
export MINIO_ACCESS_KEY=FJDSJ
export MINIO_SECRET_KEY=DSG643HGDS

mkdir -p /tmp/minio
minio server --quiet /tmp/minio &
sleep 5
go test github.com/ctrox/csi-s3-driver/pkg/s3 -cover
