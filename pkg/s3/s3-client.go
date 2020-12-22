package s3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
)

const (
	metadataName = ".metadata.json"
)

type s3Client struct {
	cfg   *Config
	minio *minio.Client
}

type volume struct {
	ID       string
	Bucket   string
	Prefix   string
	Capacity int64
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
		Mounter:         secrets["mounter"],
		CommonBucket:    secrets["commonBucket"],
		CommonPrefix:    secrets["commonPrefix"],
	})
}

func (client *s3Client) mounter() (Mounter, error) {
	return newMounter(client.cfg, client.cfg.Mounter)
}

func (client *s3Client) volumeExists(vol *volume) (bool, error) {
	client.completeVolume(vol)
	glog.Info("check volume", "volume", vol)

	if exists, err := client.minio.BucketExists(vol.Bucket); err != nil || !exists {
		return exists, err
	}

	obj, err := client.minio.GetObject(vol.Bucket, path.Join(vol.Prefix, metadataName), minio.GetObjectOptions{})
	if err != nil {
		return false, err
	}
	content := &volume{}
	if e := json.NewDecoder(obj).Decode(content); e != nil {
		return false, e
	}
	return content.ID == vol.ID && content.Bucket == vol.Bucket && content.Prefix == vol.Prefix, nil
}

func (client *s3Client) createVolume(vol *volume) error {
	client.completeVolume(vol)

	exists, err := client.minio.BucketExists(vol.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := client.minio.MakeBucket(vol.Bucket, client.cfg.Region); err != nil {
			return err
		}
	}

	b := new(bytes.Buffer)
	if e := json.NewEncoder(b).Encode(vol); e != nil {
		return e
	}
	if _, err = client.minio.PutObject(vol.Bucket, path.Join(vol.Prefix, metadataName),
		b, int64(b.Len()),
		minio.PutObjectOptions{ContentType: "application/json"}); err != nil {
		return fmt.Errorf("create metadata object: %w", err)
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

		for object := range client.minio.ListObjects(vol.Bucket, vol.Prefix, true, doneCh) {
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
		errorCh := client.minio.RemoveObjects(vol.Bucket, objectsCh)
		for e := range errorCh {
			glog.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
		}
		if len(errorCh) != 0 {
			return fmt.Errorf("Failed to remove all objects of bucket %s", vol.Bucket)
		}
	}

	if vol.Prefix != "" {
		return client.minio.RemoveObject(vol.Bucket, vol.Prefix)
	}
	return client.minio.RemoveBucket(vol.Bucket)
}

func (client *s3Client) completeVolume(vol *volume) {
	if vol.Bucket != "" {
		return
	}

	hash := uuid.NewSHA1(uuid.Nil, []byte(vol.ID)).String()

	if client.cfg.CommonBucket != "" {
		vol.Bucket = client.cfg.CommonBucket
		vol.Prefix = path.Join(client.cfg.CommonPrefix, hash)
		return
	}

	vol.Bucket = bucketNamePrefix + hash
	vol.Prefix = client.cfg.CommonPrefix
}

func (client *s3Client) getVolume(vol *volume) error {
	client.completeVolume(vol)
	obj, err := client.minio.GetObject(vol.Bucket, path.Join(vol.Prefix, metadataName), minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	return json.NewDecoder(obj).Decode(vol)
}
