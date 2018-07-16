package s3

import (
	"fmt"
	"os"
	"os/exec"
)

func s3fsMount(bucket string, cfg *Config, targetPath string) error {
	if err := writes3fsPass(cfg); err != nil {
		return err
	}
	args := []string{
		fmt.Sprintf("%s", bucket),
		fmt.Sprintf("%s", targetPath),
		"-o", "sigv2",
		"-o", "use_path_request_style",
		"-o", fmt.Sprintf("url=%s", cfg.Endpoint),
		"-o", fmt.Sprintf("endpoint=%s", cfg.Region),
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

func writes3fsPass(cfg *Config) error {
	pwFileName := fmt.Sprintf("%s/.passwd-s3fs", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(cfg.AccessKeyID + ":" + cfg.SecretAccessKey)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
