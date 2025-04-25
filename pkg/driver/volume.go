package driver

import (
	"github.com/ctrox/csi-s3/pkg/mounter"
	"github.com/golang/glog"
)

type Volume struct {
	VolumeId string

	// volume's real mount point
	stagingTargetPath string

	// Target paths to which the volume has been published.
	// These paths are symbolic links to the real mount point.
	// So multiple pods using the same volume can share a mount.
	targetPaths map[string]bool

	mounter mounter.Mounter
}

func NewVolume(volumeID string, mounter mounter.Mounter) *Volume {
	return &Volume{
		VolumeId:    volumeID,
		mounter:     mounter,
		targetPaths: make(map[string]bool),
	}
}

func (vol *Volume) Stage(stagingTargetPath string) error {
	if vol.isStaged() {
		return nil
	}

	if err := vol.mounter.Stage(stagingTargetPath); err != nil {
		return err
	}

	vol.stagingTargetPath = stagingTargetPath
	return nil
}

func (vol *Volume) Publish(targetPath string) error {
	if err := vol.mounter.Mount(vol.stagingTargetPath, targetPath); err != nil {
		return err
	}

	vol.targetPaths[targetPath] = true
	return nil
}

func (vol *Volume) Unpublish(targetPath string) error {
	// Check whether the volume is published to the target path.
	if _, ok := vol.targetPaths[targetPath]; !ok {
		glog.Warningf("volume %s hasn't been published to %s", vol.VolumeId, targetPath)
		return nil
	}

	if err := vol.mounter.Unmount(targetPath); err != nil {
		return err
	}

	delete(vol.targetPaths, targetPath)
	return nil
}

func (vol *Volume) Unstage(_ string) error {
	if !vol.isStaged() {
		return nil
	}

	if err := vol.mounter.Unstage(vol.stagingTargetPath); err != nil {
		return err
	}

	return nil
}

func (vol *Volume) isStaged() bool {
	return vol.stagingTargetPath != ""
}
