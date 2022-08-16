package mounter

import (
	"fmt"
	"os"
	"path"

	"context"

	"github.com/ctrox/csi-s3/pkg/s3"
	goofysApi "github.com/kahing/goofys/api"
	"github.com/kahing/goofys/api/common"
)

const (
	goofysCmd     = "goofys"
	defaultRegion = "us-east-1"
)

// Implements Mounter
type goofysMounter struct {
	meta            *s3.FSMeta
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

func newGoofysMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	region := cfg.Region
	// if endpoint is set we need a default region
	if region == "" && cfg.Endpoint != "" {
		region = defaultRegion
	}
	return &goofysMounter{
		meta:            meta,
		endpoint:        cfg.Endpoint,
		region:          region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
	}, nil
}

func (goofys *goofysMounter) Stage(stageTarget string) error {
	return nil
}

func (goofys *goofysMounter) Unstage(stageTarget string) error {
	return nil
}

func (goofys *goofysMounter) Mount(source string, target string) error {
	goofysCfg := &common.FlagStorage{
		MountPoint: target,
		Endpoint:   goofys.endpoint,
		DirMode:    0755,
		FileMode:   0644,
		MountOptions: map[string]string{
			"allow_other": "",
		},
		Backend: &common.S3Config{
			Region: goofys.region,
		},
	}

	os.Setenv("AWS_ACCESS_KEY_ID", goofys.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", goofys.secretAccessKey)
	fullPath := fmt.Sprintf("%s:%s", goofys.meta.BucketName, path.Join(goofys.meta.Prefix, goofys.meta.FSPath))

	_, _, err := goofysApi.Mount(context.Background(), fullPath, goofysCfg)

	if err != nil {
		return fmt.Errorf("Error mounting via goofys: %s", err)
	}
	return nil
}

func (goofys *goofysMounter) Unmount(target string) error {
	return FuseUnmount(target)
}
