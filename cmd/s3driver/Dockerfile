FROM debian:stretch
LABEL maintainers="Cyrill Troxler <cyrilltroxler@gmail.com>"
LABEL description="s3 fuse csi plugin"

RUN apt-get update && apt-get install -y s3fs && rm -rf /var/lib/apt/lists/*

COPY ./_output/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
