package s3

import (
	"fmt"
	"os"
	"os/exec"
)

func s3fsMount(bucket string, cr *Credentials, targetPath string) error {
	if err := writes3fsPass(cr); err != nil {
		return err
	}
	args := []string{
		fmt.Sprintf("%s", bucket),
		fmt.Sprintf("%s", targetPath),
		"-o", "sigv2",
		"-o", "use_path_request_style",
		"-o", fmt.Sprintf("url=%s", cr.Endpoint),
		"-o", fmt.Sprintf("endpoint=%s", cr.Region),
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

func writes3fsPass(cr *Credentials) error {
	pwFileName := fmt.Sprintf("%s/.passwd-s3fs", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(cr.AccessKeyID + ":" + cr.SecretAccessKey)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
