package s3

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/minio/minio-go"
)

type s3Client struct {
	cfg   *Config
	minio *minio.Client
}

type volume struct {
	id     string
	bucket string
	prefix string
}

func newS3Client(cfg *Config) (*s3Client, error) {
	var client = &s3Client{}

	client.cfg = cfg
	u, err := url.Parse(client.cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	ssl := u.Scheme == "https"
	endpoint := u.Hostname()
	if u.Port() != "" {
		endpoint = u.Hostname() + ":" + u.Port()
	}
	minioClient, err := minio.NewWithRegion(endpoint, client.cfg.AccessKeyID, client.cfg.SecretAccessKey, ssl, client.cfg.Region)
	if err != nil {
		return nil, err
	}
	client.minio = minioClient
	return client, nil
}

func newS3ClientFromSecrets(secrets map[string]string) (*s3Client, error) {
	return newS3Client(&Config{
		AccessKeyID:     secrets["accessKeyID"],
		SecretAccessKey: secrets["secretAccessKey"],
		Region:          secrets["region"],
		Endpoint:        secrets["endpoint"],
		// Mounter is set in the volume preferences, not secrets
		Mounter: "",
	})
}

func (client *s3Client) mounter(mounter string) (Mounter, error) {
	return newMounter(client.cfg, mounter)
}

func (client *s3Client) volumeExists(vol *volume) (bool, error) {
	client.completeVolume(vol)

	if exists, err := client.minio.BucketExists(vol.bucket); err != nil || !exists {
		return exists, err
	}
	if vol.prefix != "" {
		_, err := client.minio.GetObject(vol.bucket, vol.prefix+"/", minio.GetObjectOptions{})
		if err != nil {
			return false, err
		}
	}

	return true, nil

}

func (client *s3Client) createVolume(vol *volume) error {
	client.completeVolume(vol)

	exists, err := client.minio.BucketExists(vol.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := client.minio.MakeBucket(vol.bucket, client.cfg.Region); err != nil {
			return err
		}
	}

	if vol.prefix != "" {
		_, err := client.minio.PutObject(vol.bucket, vol.prefix+"/", bytes.NewReader([]byte{}), 0, minio.PutObjectOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (client *s3Client) removeVolume(vol *volume) error {
	client.completeVolume(vol)

	objectsCh := make(chan string)
	var listErr error

	go func() {
		defer close(objectsCh)

		doneCh := make(chan struct{})

		defer close(doneCh)

		for object := range client.minio.ListObjects(vol.bucket, vol.prefix, true, doneCh) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			objectsCh <- object.Key
		}
	}()

	if listErr != nil {
		glog.Error("Error listing objects", listErr)
		return listErr
	}

	select {
	default:
		errorCh := client.minio.RemoveObjects(vol.bucket, objectsCh)
		for e := range errorCh {
			glog.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
		}
		if len(errorCh) != 0 {
			return fmt.Errorf("Failed to remove all objects of bucket %s", vol.bucket)
		}
	}

	if vol.prefix != "" {
		return client.minio.RemoveObject(vol.bucket, vol.prefix)
	}
	return client.minio.RemoveBucket(vol.bucket)
}

func (client *s3Client) completeVolume(vol *volume) {
	if vol.bucket != "" {
		return
	}

	if client.cfg.CommonBucket != "" {
		vol.bucket = client.cfg.CommonBucket
		vol.prefix = filepath.Join(client.cfg.CommonPrefix, vol.id)
		return
	}

	vol.bucket = bucketNamePrefix + vol.id
	vol.prefix = client.cfg.CommonPrefix
}
