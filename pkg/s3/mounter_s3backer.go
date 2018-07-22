package s3

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/util/mount"
)

// Implements Mounter
type s3backerMounter struct {
	bucket          string
	url             string
	region          string
	accessKeyID     string
	secretAccessKey string
	size            int64
	initMountPath   string
}

const (
	s3backerCmd       = "s3backer"
	s3backerFsType    = "xfs"
	s3backerMountBase = "/mnt"
	s3backerDevice    = "file"
	// blockSize to use in k
	s3backerBlockSize = "128k"
)

func newS3backerMounter(bucket string, cfg *Config) (Mounter, error) {
	s3backer := &s3backerMounter{
		bucket:          bucket,
		url:             cfg.Endpoint,
		region:          cfg.Region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
		initMountPath:   path.Join(s3backerMountBase, bucket),
		size:            1024 * 1024 * 1024 * 10,
	}

	return s3backer, s3backer.writePasswd()
}

func (s3backer *s3backerMounter) String() string {
	return s3backer.bucket
}

func (s3backer *s3backerMounter) Format() error {
	tmpDir, err := ioutil.TempDir("", "s3backer")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := s3backer.mountInit(tmpDir); err != nil {
		return err
	}
	defer fuseUnmount(tmpDir, s3backerCmd)

	return formatFs(s3backerFsType, path.Join(tmpDir, s3backerDevice))
}

func (s3backer *s3backerMounter) Mount(targetPath string) error {
	if err := os.MkdirAll(s3backer.initMountPath, 0700); err != nil {
		return err
	}
	// s3backer requires two mounts
	// first mount will fuse mount the bucket to a single 'file'
	err := s3backer.mountInit(s3backer.initMountPath)
	if err != nil {
		return err
	}
	device := path.Join(s3backer.initMountPath, s3backerDevice)
	// second mount will mount the 'file' as a filesystem
	err = mount.New("").Mount(device, targetPath, s3backerFsType, []string{})
	if err != nil {
		// cleanup fuse mount
		fuseUnmount(targetPath, s3backerCmd)
		return err
	}
	return nil
}

func (s3backer *s3backerMounter) Unmount(targetPath string) error {
	// Unmount the filesystem first
	if err := mount.New("").Unmount(targetPath); err != nil {
		return err
	}
	// Unmount the s3backer fuse mount
	err := fuseUnmount(s3backer.initMountPath, s3backerCmd)
	if err != nil {
		return err
	}
	return nil
}

func (s3backer *s3backerMounter) mountInit(path string) error {
	args := []string{
		// baseURL must end with /
		fmt.Sprintf("--baseURL=%s/", s3backer.url),
		fmt.Sprintf("--blockSize=%v", s3backerBlockSize),
		fmt.Sprintf("--size=%v", s3backer.size),
		"--listBlocks",
		s3backer.bucket,
		path,
	}
	if s3backer.region != "" {
		args = append(args, fmt.Sprintf("--region=%s", s3backer.region))
	}

	return fuseMount(path, s3backerCmd, args)
}

func (s3backer *s3backerMounter) writePasswd() error {
	pwFileName := fmt.Sprintf("%s/.s3backer_passwd", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(s3backer.accessKeyID + ":" + s3backer.secretAccessKey)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}

func formatFs(fsType string, device string) error {
	diskMounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}
	format, err := diskMounter.GetDiskFormat(device)
	if err != nil {
		return err
	}
	if format != "" {
		glog.Infof("Disk %s is already formatted with format %s", device, format)
		return nil
	}
	args := []string{
		device,
	}
	cmd := exec.Command("mkfs."+fsType, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error formatting disk: %s", out)
	}
	glog.Info("Formatting fs with type %s, out: %s", fsType, out)
	return nil
}
