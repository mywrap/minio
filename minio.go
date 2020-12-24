package minio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	minio "github.com/minio/minio-go/v6"
	"github.com/mywrap/log"
)

// Client connects to a MinIO server, create a predefined bucketName if needed,
// then upload a test file. This client only uploads to the bucketName.
type Client struct {
	client *minio.Client
	bucket string
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.BucketName == "" {
		return nil, errors.New("have to define a default bucket")
	}

	endpoint := fmt.Sprintf("%v:%v", cfg.EndpointHost, cfg.EndpointPort)
	cli, err := minio.New(endpoint, cfg.AccessId, cfg.AccessSecret, cfg.IsEndpointTLS)
	if err != nil {
		return nil, err
	}

	// create the bucketName if needed and set its policy public read
	ctx, cxl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cxl()
	err = cli.MakeBucketWithContext(ctx, cfg.BucketName, "")
	if err != nil {
		isExisted, err2 := cli.BucketExistsWithContext(ctx, cfg.BucketName)
		if err2 != nil {
			return nil, fmt.Errorf("check bucket existed: %v", err2)
		}
		if !isExisted {
			return nil, fmt.Errorf("create bucket: %v", err2)
		}
	} else {
		log.Printf("created MinIO bucket %v\n", cfg.BucketName)
		// https://docs.min.io/docs/golang-client-api-reference#SetBucketPolicy
		policy := fmt.Sprintf(`{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": ["s3:GetObject"],
			"Effect": "Allow",
			"Principal": {
				"AWS": ["*"]
			},
			"Resource": [
				"arn:aws:s3:::%v/*"
			],
			"Sid": "PublicReadGetObject"
		}
	]
}`, cfg.BucketName)
		err = cli.SetBucketPolicy(cfg.BucketName, policy)
		if err != nil {
			return nil, fmt.Errorf("set bucket policy: %v", err)
		}
	}

	myClient := &Client{client: cli, bucket: cfg.BucketName}
	_, err = myClient.UploadWithCtx(ctx, "",
		"PING", []byte(fmt.Sprintf("PING at %v", time.Now())))
	if err != nil {
		return nil, fmt.Errorf("test upload: %v", err)
	}
	log.Printf("successfully inited Client to %v", endpoint)
	return myClient, nil
}

// uploadWithCtx uploads input file to the client's predefined bucket.
// if the file uploaded successfully, people can read it at URL:
// 		{host}:{port}/{bucketName}/{fileName}
//		ex: http://127.0.0.1:9000/bucket0/DDCat.jpg
// :param contentType: default is "text/plain;charset=UTF-8", detail:
// 		https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types
func (c Client) UploadWithCtx(ctx context.Context, contentType string,
	fileName string, data []byte) (string, error) {
	objSize := int64(len(data))
	if contentType == "" {
		contentType = "text/plain;charset=UTF-8"
	}
	option := minio.PutObjectOptions{ContentType: contentType}
	n, err := c.client.PutObjectWithContext(ctx, c.bucket, fileName,
		bytes.NewReader(data), objSize, option)
	if err != nil {
		return "", fmt.Errorf("client PutObjectWithContext: %v", err)
	}
	if n != objSize {
		return "", fmt.Errorf("size: expected %v, real %v", objSize, n)
	}
	return c.GetPath(fileName), nil
}

// Upload is for testing,
// Deprecated: use uploadWithCtx for more control
func (c Client) Upload(fileName string, data []byte) error {
	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()
	_, err := c.UploadWithCtx(ctx, "", fileName, data)
	return err
}

func (c Client) GetPath(fileName string) string {
	return fmt.Sprintf("/%v/%v", c.bucket, fileName)
}

// Config can be loaded easily by calling func LoadEnvConfig
type Config struct {
	EndpointHost  string
	EndpointPort  string
	IsEndpointTLS bool
	AccessId      string
	AccessSecret  string
	BucketName    string
}

// LoadEnvConfig loads config from environment variables:
// MINIO_HOST, MINIO_PORT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET_NAME
func LoadEnvConfig() Config {
	isEndpointTLS, _ := strconv.ParseBool(os.Getenv("MINIO_IS_TLS"))
	return Config{
		EndpointHost:  os.Getenv("MINIO_HOST"),
		EndpointPort:  os.Getenv("MINIO_PORT"),
		IsEndpointTLS: isEndpointTLS,
		AccessId:      os.Getenv("MINIO_ACCESS_KEY"),
		AccessSecret:  os.Getenv("MINIO_SECRET_KEY"),
		BucketName:    os.Getenv("MINIO_BUCKET_NAME"),
	}
}
