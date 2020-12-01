package s3

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/util/mount"
)

// Mounter interface which can be implemented
// by the different mounter types
type Mounter interface {
	Stage(vol *volume, stagePath string) error
	Unstage(vol *volume, stagePath string) error
	Mount(vol *volume, source string, target string) error
}

const (
	s3fsMounterType     = "s3fs"
	goofysMounterType   = "goofys"
	s3backerMounterType = "s3backer"
	rcloneMounterType   = "rclone"
	mounterTypeKey      = "mounter"
)

// newMounter returns a new mounter depending on the mounterType parameter
func newMounter(cfg *Config, mounter string) (Mounter, error) {
	// Fall back to mounterType in cfg
	if len(mounter) == 0 {
		mounter = cfg.Mounter
	}
	switch mounter {
	case s3fsMounterType:
		return newS3fsMounter(cfg)

	case goofysMounterType:
		return newGoofysMounter(cfg)

	case s3backerMounterType:
		return newS3backerMounter(cfg)

	case rcloneMounterType:
		return newRcloneMounter(cfg)

	default:
		// default to s3backer
		return newS3backerMounter(cfg)
	}
}

func fuseMount(path string, command string, args []string) error {
	cmd := exec.Command(command, args...)
	glog.V(3).Infof("Mounting fuse with command: %s and args: %s", command, args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error fuseMount command: %s\nargs: %s\noutput: %s", command, args, out)
	}

	return waitForMount(path, 10*time.Second)
}

func fuseUnmount(path string) error {
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
