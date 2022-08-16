package mounter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/ctrox/csi-s3/pkg/s3"
	"github.com/golang/glog"
	"github.com/mitchellh/go-ps"
	"k8s.io/mount-utils"
)

// Mounter interface which can be implemented
// by the different mounter types
type Mounter interface {
	Stage(stagePath string) error
	Unstage(stagePath string) error
	Mount(source string, target string) error
	Unmount(target string) error
}

const (
	s3fsMounterType     = "s3fs"
	goofysMounterType   = "goofys"
	s3backerMounterType = "s3backer"
	rcloneMounterType   = "rclone"
	TypeKey             = "mounter"
	BucketKey           = "bucket"
	VolumePrefix        = "prefix"
	UsePrefix           = "usePrefix"
)

// New returns a new mounter depending on the mounterType parameter
func New(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	mounter := meta.Mounter
	// Fall back to mounterType in cfg
	if len(meta.Mounter) == 0 {
		mounter = cfg.Mounter
	}
	switch mounter {
	case s3fsMounterType:
		return newS3fsMounter(meta, cfg)

	case goofysMounterType:
		return newGoofysMounter(meta, cfg)

	case s3backerMounterType:
		return newS3backerMounter(meta, cfg)

	case rcloneMounterType:
		return newRcloneMounter(meta, cfg)

	default:
		// default to s3backer
		return newS3backerMounter(meta, cfg)
	}
}

func fuseMount(path string, command string, args []string) error {
	cmd := exec.Command(command, args...)
	glog.V(3).Infof("Mounting fuse with command: %s and args: %s", command, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error fuseMount command: %s\nargs: %s\noutput", command, args)
	}

	return waitForMount(path, 10*time.Second)
}

func FuseUnmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	// as fuse quits immediately, we will try to wait until the process is done
	process, err := findFuseMountProcess(path)
	if err != nil {
		glog.Errorf("Error getting PID of fuse mount: %s", err)
		return nil
	}
	if process == nil {
		glog.Warningf("Unable to find PID of fuse mount %s, it must have finished already", path)
		return nil
	}
	glog.Infof("Found fuse pid %v of mount %s, checking if it still runs", process.Pid, path)
	return waitForProcess(process, 1)
}

func waitForMount(path string, timeout time.Duration) error {
	var elapsed time.Duration
	var interval = 10 * time.Millisecond
	for {
		notMount, err := mount.New("").IsLikelyNotMountPoint(path)
		if err != nil {
			return err
		}
		if !notMount {
			return nil
		}
		time.Sleep(interval)
		elapsed = elapsed + interval
		if elapsed >= timeout {
			return errors.New("Timeout waiting for mount")
		}
	}
}

func findFuseMountProcess(path string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		cmdLine, err := getCmdLine(p.Pid())
		if err != nil {
			glog.Errorf("Unable to get cmdline of PID %v: %s", p.Pid(), err)
			continue
		}
		if strings.Contains(cmdLine, path) {
			glog.Infof("Found matching pid %v on path %s", p.Pid(), path)
			return os.FindProcess(p.Pid())
		}
	}
	return nil, nil
}

func waitForProcess(p *os.Process, backoff int) error {
	if backoff == 20 {
		return fmt.Errorf("Timeout waiting for PID %v to end", p.Pid)
	}
	cmdLine, err := getCmdLine(p.Pid)
	if err != nil {
		glog.Warningf("Error checking cmdline of PID %v, assuming it is dead: %s", p.Pid, err)
		return nil
	}
	if cmdLine == "" {
		// ignore defunct processes
		// TODO: debug why this happens in the first place
		// seems to only happen on k8s, not on local docker
		glog.Warning("Fuse process seems dead, returning")
		return nil
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		glog.Warningf("Fuse process does not seem active or we are unprivileged: %s", err)
		return nil
	}
	glog.Infof("Fuse process with PID %v still active, waiting...", p.Pid)
	time.Sleep(time.Duration(backoff*100) * time.Millisecond)
	return waitForProcess(p, backoff+1)
}

func getCmdLine(pid int) (string, error) {
	cmdLineFile := fmt.Sprintf("/proc/%v/cmdline", pid)
	cmdLine, err := ioutil.ReadFile(cmdLineFile)
	if err != nil {
		return "", err
	}
	return string(cmdLine), nil
}

func createLoopDevice(device string) error {
	if _, err := os.Stat(device); !os.IsNotExist(err) {
		return nil
	}
	args := []string{
		device,
		"b", "7", "0",
	}
	cmd := exec.Command("mknod", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error configuring loop device: %s", out)
	}
	return nil
}
