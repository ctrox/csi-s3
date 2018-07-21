package s3

import (
	"fmt"
	"os"
	"os/exec"
)

// Implements Mounter
type s3fsMounter struct {
	bucket        string
	url           string
	region        string
	pwFileContent string
}

func newS3fsMounter(bucket string, cfg *Config) (Mounter, error) {
	return &s3fsMounter{
		bucket:        bucket,
		url:           cfg.Endpoint,
		region:        cfg.Region,
		pwFileContent: cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (s3fs *s3fsMounter) Format() error {
	return nil
}

func (s3fs *s3fsMounter) Mount(targetPath string) error {
	if err := writes3fsPass(s3fs.pwFileContent); err != nil {
		return err
	}
	args := []string{
		fmt.Sprintf("%s", s3fs.bucket),
		fmt.Sprintf("%s", targetPath),
		"-o", "sigv2",
		"-o", "use_path_request_style",
		"-o", fmt.Sprintf("url=%s", s3fs.url),
		"-o", fmt.Sprintf("endpoint=%s", s3fs.region),
		"-o", "allow_other",
		"-o", "mp_umask=000",
	}
	cmd := exec.Command("s3fs", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error mounting using s3fs, output: %s", out)
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
