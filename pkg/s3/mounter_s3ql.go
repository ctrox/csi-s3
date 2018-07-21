package s3

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/ini.v1"
)

// Implements Mounter
type s3qlMounter struct {
	url        string
	bucketURL  string
	login      string
	password   string
	passphrase string
	options    []string
	ssl        bool
	targetPath string
}

const (
	s3qlCmdMkfs  = "mkfs.s3ql"
	s3qlCmdMount = "mount.s3ql"
)

func newS3qlMounter(bucket string, cfg *Config) (Mounter, error) {
	url, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	ssl := url.Scheme != "http"
	if strings.Contains(url.Scheme, "http") {
		url.Scheme = "s3c"
	}
	s3ql := &s3qlMounter{
		url:        url.String(),
		login:      cfg.AccessKeyID,
		password:   cfg.SecretAccessKey,
		passphrase: cfg.EncryptionKey,
		ssl:        ssl,
	}

	url.Path = path.Join(url.Path, bucket)
	s3ql.bucketURL = url.String()

	if !ssl {
		s3ql.options = []string{"--backend-options", "no-ssl"}
	}

	return s3ql, s3ql.writeConfig()
}

func (s3ql *s3qlMounter) Format() error {
	// force creation to ignore existing data
	args := []string{
		s3ql.bucketURL,
		"--force",
	}

	p := fmt.Sprintf("%s\n%s\n", s3ql.passphrase, s3ql.passphrase)
	reader := bytes.NewReader([]byte(p))
	return s3qlCmd(s3qlCmdMkfs, append(args, s3ql.options...), reader)
}

func (s3ql *s3qlMounter) Mount(targetPath string) error {
	args := []string{
		s3ql.bucketURL,
		targetPath,
		"--allow-other",
	}
	return s3qlCmd(s3qlCmdMount, append(args, s3ql.options...), nil)
}

func s3qlCmd(s3qlCmd string, args []string, stdin io.Reader) error {
	cmd := exec.Command(s3qlCmd, args...)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running s3ql command: %s", out)
	}
	return nil
}

func (s3ql *s3qlMounter) writeConfig() error {
	s3qlIni := ini.Empty()
	section, err := s3qlIni.NewSection("s3ql")
	if err != nil {
		return err
	}

	section.NewKey("storage-url", s3ql.url)
	section.NewKey("backend-login", s3ql.login)
	section.NewKey("backend-password", s3ql.password)
	section.NewKey("fs-passphrase", s3ql.passphrase)

	authDir := os.Getenv("HOME") + "/.s3ql"
	authFile := authDir + "/authinfo2"
	os.Mkdir(authDir, 0700)
	s3qlIni.SaveTo(authFile)
	os.Chmod(authFile, 0600)
	return nil
}
