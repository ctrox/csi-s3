module github.com/ctrox/csi-s3

go 1.15

require (
	github.com/Azure/azure-sdk-for-go v32.1.0+incompatible // indirect
	github.com/Azure/azure-storage-blob-go v0.7.1-0.20190724222048-33c102d4ffd2 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.11 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/aws/aws-sdk-go v1.42.44 // indirect
	github.com/container-storage-interface/spec v1.1.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/jacobsa/fuse v0.0.0-00010101000000-000000000000 // indirect
	github.com/kahing/goofys v0.24.0
	github.com/kubernetes-csi/csi-lib-utils v0.6.1 // indirect
	github.com/kubernetes-csi/csi-test v2.0.0+incompatible
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/minio/minio-go/v7 v7.0.5
	github.com/mitchellh/go-ps v0.0.0-20170309133038-4fdf99ab2936
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/shirou/gopsutil v2.21.11+incompatible // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/urfave/cli v1.22.5 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/net v0.0.0-20211216030914-fe4d6282115f
	google.golang.org/genproto v0.0.0-20220126215142-9970aeb2e350 // indirect
	google.golang.org/grpc v1.40.0
	k8s.io/mount-utils v0.23.3
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
)

replace github.com/jacobsa/fuse => github.com/kahing/fusego v0.0.0-20200327063725-ca77844c7bcc
