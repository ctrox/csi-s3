FROM golang:stretch
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="s3 fuse csi plugin"

RUN apt-get update && apt-get install -y s3fs && rm -rf /var/lib/apt/lists/*
RUN go get -u github.com/minio/minio && go install github.com/minio/minio/cmd

CMD ["/go/src/github.com/ctrox/csi-s3-driver/test/test.sh"]
