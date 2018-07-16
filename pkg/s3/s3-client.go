package s3

import (
	"fmt"
	"net/url"

	"github.com/golang/glog"
	"github.com/minio/minio-go"
)

const (
	metadataName = ".meta"
)

type s3Client struct {
	cfg   *Config
	minio *minio.Client
}

type bucketMetadata struct {
	CapacityBytes int64
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
	minioClient, err := minio.New(endpoint, client.cfg.AccessKeyID, client.cfg.SecretAccessKey, ssl)
	if err != nil {
		return nil, err
	}
	client.minio = minioClient
	return client, nil
}

func (client *s3Client) bucketExists(bucketName string) (bool, error) {
	return client.minio.BucketExists(bucketName)
}

func (client *s3Client) createBucket(bucketName string) error {
	return client.minio.MakeBucket(bucketName, client.cfg.Region)
}

func (client *s3Client) removeBucket(bucketName string) error {
	if err := client.emptyBucket(bucketName); err != nil {
		return err
	}
	return client.minio.RemoveBucket(bucketName)
}

func (client *s3Client) emptyBucket(bucketName string) error {
	objectsCh := make(chan string)
	var listErr error

	go func() {
		defer close(objectsCh)

		doneCh := make(chan struct{})

		defer close(doneCh)

		for object := range client.minio.ListObjects(bucketName, "", true, doneCh) {
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
		errorCh := client.minio.RemoveObjects(bucketName, objectsCh)
		for e := range errorCh {
			glog.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
		}
		if len(errorCh) != 0 {
			return fmt.Errorf("Failed to remove all objects of bucket %s", bucketName)
		}
	}

	return nil
}
