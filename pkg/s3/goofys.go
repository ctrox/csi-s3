package s3

import (
	"fmt"
	"os"

	"context"

	goofys "github.com/kahing/goofys/api"
)

const defaultRegion = "us-east-1"

func goofysMount(bucket string, cfg *Config, targetPath string) error {
	goofysCfg := &goofys.Config{
		MountPoint: targetPath,
		Endpoint:   cfg.Endpoint,
		Region:     cfg.Region,
		DirMode:    0755,
		FileMode:   0644,
		MountOptions: map[string]string{
			"allow_other": "",
		},
	}
	if cfg.Endpoint != "" {
		cfg.Region = defaultRegion
	}
	os.Setenv("AWS_ACCESS_KEY_ID", cfg.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", cfg.SecretAccessKey)

	_, _, err := goofys.Mount(context.Background(), bucket, goofysCfg)

	if err != nil {
		return fmt.Errorf("Error mounting via goofys: %s", err)
	}
	return nil
}
