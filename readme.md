# MinIO client

MinIO is a high performance object storage, open source, Amazon S3
compatible, Kubernetes Native and is designed for cloud native
workloads like AI.

This client wrapped [github.com/minio](
https://github.com/minio/minio-go/tree/v6.0.57) Go client.

## Usage

````go
// MINIO_HOST, MINIO_PORT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET_NAME
conf := LoadEnvConfig()
client, err := NewClient(conf)
if err != nil {
    t.Fatalf("error init MinIOClient: %v, config: %#v", err, conf)
}

// upload a small file
	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()
	err = client.UploadWithCtx(ctx, "text/plain;charset=UTF-8",
		"TestNewClient", []byte("puss"))
	if err != nil {
		t.Error(err)
	}
````
Detail in [minio_test.go](./minio_test.go).
