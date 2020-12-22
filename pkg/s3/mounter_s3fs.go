package s3

import (
	"fmt"
	"os"
	"strings"
)

// Implements Mounter
type s3fsMounter struct {
	url           string
	region        string
	pwFileContent string
}

const (
	s3fsCmd = "s3fs"
)

func newS3fsMounter(cfg *Config) (Mounter, error) {
	return &s3fsMounter{
		url:           cfg.Endpoint,
		region:        cfg.Region,
		pwFileContent: cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (s3fs *s3fsMounter) Stage(*volume, string) error {
	return nil
}

func (s3fs *s3fsMounter) Unstage(*volume, string) error {
	return nil
}

func (s3fs *s3fsMounter) Mount(vol *volume, source string, target string) error {
	if err := writes3fsPass(s3fs.pwFileContent); err != nil {
		return fmt.Errorf("write s3 fs pass: %w", err)
	}

	dev := vol.Bucket
	if vol.Prefix != "" {
		dev += ":/" + strings.TrimSuffix(vol.Prefix, "/")
	}

	opts := []string{
		dev,
		target,
		"-f",
		"-o", "use_path_request_style",
		"-o", "url=" + s3fs.url,
		"-o", "endpoint=" + s3fs.region,
		"-o", "allow_other",
		"-o", "mp_umask=000",
	}

	if e := fuseMount(target, s3fsCmd, opts); e != nil {
		return fmt.Errorf("failed to mount s3fs: %w", e)
	}

	return nil
}

func writes3fsPass(pwFileContent string) error {
	pwFileName := fmt.Sprintf("%s/.passwd-s3fs", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(pwFileContent)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
