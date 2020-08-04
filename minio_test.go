package minio

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

// Need to setup a MinIO server and set env to run this test.
// Setup a server: github.com/daominah/docker/minio
func TestNewClient(t *testing.T) {

	// connect to server
	conf := LoadEnvConfig()
	client, err := NewClient(conf)
	if err != nil {
		t.Fatalf("error init MinIOClient: %v, config: %#v", err, conf)
	}

	// upload a small file
	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()
	_, err = client.UploadWithCtx(ctx, "",
		"TestNewClient", []byte("pussy"))
	if err != nil {
		t.Error(err)
	}

	// upload a 10MB file
	lines := make([]string, 10240)
	for i := 0; i < len(lines); i++ {
		lines[i] = fmt.Sprintf("lines %05d: %v\n",
			i+1, strings.Repeat("0123456789", 101))
	}
	bigFile := []byte(strings.Join(lines, ""))
	err = client.Upload("TestBigFile", bigFile)
	if err != nil {
		t.Error(err)
	}

	// upload an image
	img, _ := ioutil.ReadFile("./DDCat_test.jpg")
	_, err = client.UploadWithCtx(context.Background(), "image/jpeg",
		"DDCat.jpg", img)
	if err != nil {
		t.Error(err)
	}

	// create directory
	_, err = client.UploadWithCtx(context.Background(), "",
		"dir_test/hihi.txt", []byte(time.Now().Format(time.RFC3339)))
	if err != nil {
		t.Error(err)
	}

	// try invalid config
	invalidConf := Config{}
	_, err = NewClient(invalidConf)
	if err == nil {
		t.Error("expected invalid config error")
	}

	// try timeout
	ctx2, cxl2 := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cxl2()
	_, err = client.UploadWithCtx(ctx2, "", "TestTimeout", bigFile)
	if err == nil {
		t.Error("expected context deadline exceeded")
	}
}
