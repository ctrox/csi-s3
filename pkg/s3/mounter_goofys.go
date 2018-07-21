package s3

import (
	"fmt"
	"os"

	"context"

	goofysApi "github.com/kahing/goofys/api"
	"k8s.io/kubernetes/pkg/util/mount"
)

const defaultRegion = "us-east-1"

// Implements Mounter
type goofysMounter struct {
	bucket          string
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

func newGoofysMounter(bucket string, cfg *Config) (Mounter, error) {
	region := cfg.Region
	// if endpoint is set we need a default region
	if region == "" && cfg.Endpoint != "" {
		region = defaultRegion
	}
	return &goofysMounter{
		bucket:          bucket,
		endpoint:        cfg.Endpoint,
		region:          region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
	}, nil
}

func (goofys *goofysMounter) Format() error {
	return nil
}

func (goofys *goofysMounter) Mount(targetPath string) error {
	goofysCfg := &goofysApi.Config{
		MountPoint: targetPath,
		Endpoint:   goofys.endpoint,
		Region:     goofys.region,
		DirMode:    0755,
		FileMode:   0644,
		MountOptions: map[string]string{
			"allow_other": "",
		},
	}

	os.Setenv("AWS_ACCESS_KEY_ID", goofys.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", goofys.secretAccessKey)

	_, _, err := goofysApi.Mount(context.Background(), goofys.bucket, goofysCfg)

	if err != nil {
		return fmt.Errorf("Error mounting via goofys: %s", err)
	}
	return nil
}

func (goofys *goofysMounter) Unmount(targetPath string) error {
	return mount.New("").Unmount(targetPath)
}
