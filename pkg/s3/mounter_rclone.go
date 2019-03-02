package s3

import (
	"fmt"
)

// Implements Mounter
type rcloneMounter struct {
	bucket        *bucket
	url           string
	region        string
	pwFileContent string
}

const (
	rcloneCmd = "rclone"
)

func newRcloneMounter(b *bucket, cfg *Config) (Mounter, error) {
	return &rcloneMounter{
		bucket:        b,
		url:           cfg.Endpoint,
		region:        cfg.Region,
		pwFileContent: cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (rclone *rcloneMounter) Stage(stageTarget string) error {
	return nil
}

func (rclone *rcloneMounter) Unstage(stageTarget string) error {
	return nil
}

func (rclone *rcloneMounter) Mount(source string, target string) error {
	args := []string{
		"mount",
		"--daemon",
		fmt.Sprintf(":s3:%s/%s", rclone.bucket.Name, rclone.bucket.FSPath),
		fmt.Sprintf("%s", target),
		"--s3-provider=AWS",
		"--s3-env-auth=true",
		fmt.Sprintf("--s3-region=%s", rclone.region),
		fmt.Sprintf("--s3-endpoint=%s", rclone.url),
		"--allow-other",
		"--vfs-cache-mode", "minimal",
	}
	return fuseMount(target, rcloneCmd, args)
}

func (rclone *rcloneMounter) Unmount(target string) error {
	return fuseUnmount(target, rcloneCmd)
}
