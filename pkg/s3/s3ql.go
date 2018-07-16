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

type s3fsConfig struct {
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

func newS3ql(bucket string, targetPath string, cfg *Config) (*s3fsConfig, error) {
	url, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	ssl := url.Scheme != "http"
	if strings.Contains(url.Scheme, "http") {
		url.Scheme = "s3c"
	}
	s3ql := &s3fsConfig{
		url:        url.String(),
		login:      cfg.AccessKeyID,
		password:   cfg.SecretAccessKey,
		passphrase: cfg.EncryptionKey,
		ssl:        ssl,
		targetPath: targetPath,
	}

	url.Path = path.Join(url.Path, bucket)
	s3ql.bucketURL = url.String()

	if !ssl {
		s3ql.options = []string{"--backend-options", "no-ssl"}
	}

	return s3ql, s3ql.writeConfig()
}

func s3qlCreate(bucket string, cfg *Config) error {
	s3ql, err := newS3ql(bucket, "unknown", cfg)
	if err != nil {
		return err
	}
	return s3ql.create()
}

func s3qlMount(bucket string, cfg *Config, targetPath string) error {
	s3ql, err := newS3ql(bucket, targetPath, cfg)
	if err != nil {
		return err
	}

	return s3ql.mount()
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

func (cfg *s3fsConfig) create() error {
	// force creation to ignore existing data
	args := []string{
		cfg.bucketURL,
		"--force",
	}

	p := fmt.Sprintf("%s\n%s\n", cfg.passphrase, cfg.passphrase)
	reader := bytes.NewReader([]byte(p))
	return s3qlCmd(s3qlCmdMkfs, append(args, cfg.options...), reader)
}

func (cfg *s3fsConfig) mount() error {
	args := []string{
		cfg.bucketURL,
		cfg.targetPath,
		"--allow-other",
	}
	return s3qlCmd(s3qlCmdMount, append(args, cfg.options...), nil)
}

func (cfg *s3fsConfig) writeConfig() error {
	s3qlIni := ini.Empty()
	section, err := s3qlIni.NewSection("s3ql")
	if err != nil {
		return err
	}

	section.NewKey("storage-url", cfg.url)
	section.NewKey("backend-login", cfg.login)
	section.NewKey("backend-password", cfg.password)
	section.NewKey("fs-passphrase", cfg.passphrase)

	authDir := os.Getenv("HOME") + "/.s3ql"
	authFile := authDir + "/authinfo2"
	os.Mkdir(authDir, 0700)
	s3qlIni.SaveTo(authFile)
	os.Chmod(authFile, 0600)
	return nil
}
