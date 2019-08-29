package s3

import (
	"fmt"
	"os"
	"strconv"

	"context"

	"github.com/golang/glog"
	goofysApi "github.com/kahing/goofys/api"
)

const (
	goofysCmd     = "goofys"
	defaultRegion = "us-east-1"
	goofysUIDKey  = "goofysuid"
	goofysGIDKey  = "goofysgid"
)

// Implements Mounter
type goofysMounter struct {
	bucket          *bucket
	endpoint        string
	region          string
	accessKeyID     string
	secretAccessKey string
}

func newGoofysMounter(b *bucket, cfg *Config) (Mounter, error) {
	region := cfg.Region
	// if endpoint is set we need a default region
	if region == "" && cfg.Endpoint != "" {
		region = defaultRegion
	}
	return &goofysMounter{
		bucket:          b,
		endpoint:        cfg.Endpoint,
		region:          region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
	}, nil
}

func (goofys *goofysMounter) Stage(stageTarget string) error {
	return nil
}

func (goofys *goofysMounter) Unstage(stageTarget string) error {
	return nil
}

func (goofys *goofysMounter) Mount(source string, target string, attrib map[string]string) error {
	uid := uint32FromMap(goofysUIDKey, 0, attrib)
	gid := uint32FromMap(goofysGIDKey, 0, attrib)

	glog.V(4).Infof("target %v\nendpoint %v\nregion %v\nuid %v\ngid %v\nattrib %v\n",
		target, goofys.endpoint, goofys.region, uid, gid, attrib)

	goofysCfg := &goofysApi.Config{
		MountPoint: target,
		Endpoint:   goofys.endpoint,
		Region:     goofys.region,
		DirMode:    0755,
		FileMode:   0644,
		Uid:        uid,
		Gid:        gid,
		MountOptions: map[string]string{
			"allow_other": "",
		},
	}

	os.Setenv("AWS_ACCESS_KEY_ID", goofys.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", goofys.secretAccessKey)
	fullPath := fmt.Sprintf("%s:%s", goofys.bucket.Name, goofys.bucket.FSPath)

	_, _, err := goofysApi.Mount(context.Background(), fullPath, goofysCfg)

	if err != nil {
		return fmt.Errorf("Error mounting via goofys: %s", err)
	}
	return nil
}

func uint32FromMap(key string, defaultValue uint32, attrib map[string]string) uint32 {
	value := defaultValue
	if valueStr, found := attrib[key]; found {
		if i, err := strconv.ParseUint(valueStr, 10, 32); err == nil {
			value = uint32(i)
		}
	}
	return value
}
