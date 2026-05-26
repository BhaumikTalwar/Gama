//go:build integration

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getS3Config() (endpoint, region, accessKey, secretKey, publicBucket, privateBucket, publicURL string) {
	endpoint = getEnvDefault("TEST_S3_ENDPOINT", "http://localhost:9000")
	region = getEnvDefault("TEST_S3_REGION", "us-east-1")
	accessKey = getEnvDefault("TEST_S3_ACCESS_KEY", "minioadmin")
	secretKey = getEnvDefault("TEST_S3_SECRET_KEY", "minioadmin")
	publicBucket = getEnvDefault("TEST_S3_PUBLIC_BUCKET", "gama-public")
	privateBucket = getEnvDefault("TEST_S3_PRIVATE_BUCKET", "gama-private")
	publicURL = getEnvDefault("TEST_S3_PUBLIC_URL", "http://localhost:9000/gama-public")
	return
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ensureBucketsExist(t *testing.T) {
	t.Helper()
	endpoint, region, accessKey, secretKey, publicBucket, privateBucket, _ := getS3Config()

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	require.NoError(t, err)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	ctx := context.Background()
	for _, bucket := range []string{publicBucket, privateBucket} {
		_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
		if err != nil {
			_, createErr := client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(bucket),
			})
			if createErr != nil {
				t.Logf("note: failed to create bucket %s: %v", bucket, createErr)
			}
		}
	}
}

func setupS3Store(t *testing.T) *S3Store {
	t.Helper()
	ensureBucketsExist(t)
	endpoint, region, accessKey, secretKey, publicBucket, privateBucket, publicURL := getS3Config()
	store, err := NewS3Store(endpoint, region, accessKey, secretKey, publicBucket, privateBucket, publicURL)
	require.NoError(t, err)
	require.NotNil(t, store)
	return store
}

func TestNewS3Store_Success(t *testing.T) {
	store := setupS3Store(t)
	assert.NotNil(t, store.client)
	assert.NotNil(t, store.uploader)
	assert.NotNil(t, store.downloader)
	assert.NotNil(t, store.presigner)
	assert.Equal(t, "gama-public", store.publicBucket)
	assert.Equal(t, "gama-private", store.privateBucket)
}

func TestS3Store_UploadPublicAndDelete(t *testing.T) {
	store := setupS3Store(t)
	ctx := context.Background()
	key := "test/" + xid.New().String() + ".txt"
	content := "hello public s3"

	err := store.UploadPublic(ctx, key, strings.NewReader(content))
	require.NoError(t, err)

	err = store.DeletePublic(ctx, key)
	require.NoError(t, err)
}

func TestS3Store_UploadPrivateDownloadAndDelete(t *testing.T) {
	store := setupS3Store(t)
	ctx := context.Background()
	key := "test/" + xid.New().String() + ".txt"
	content := "hello private s3"

	err := store.UploadPrivate(ctx, key, strings.NewReader(content))
	require.NoError(t, err)

	body, err := store.DownloadStreamPrivate(ctx, key)
	require.NoError(t, err)
	defer body.Close()

	data, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	err = store.DeletePrivate(ctx, key)
	require.NoError(t, err)
}

func TestS3Store_DownloadPrivateToWriter(t *testing.T) {
	store := setupS3Store(t)
	ctx := context.Background()
	key := "test/" + xid.New().String() + ".bin"
	content := []byte("binary data for writer test")

	err := store.UploadPrivate(ctx, key, bytes.NewReader(content))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = store.DownloadPrivate(ctx, key, &writerAtAdapter{buf: &buf})
	require.NoError(t, err)
	assert.Equal(t, content, buf.Bytes())

	store.DeletePrivate(ctx, key)
}

type writerAtAdapter struct {
	buf *bytes.Buffer
}

func (w *writerAtAdapter) WriteAt(p []byte, off int64) (int, error) {
	if int64(len(w.buf.Bytes())) < off {
		w.buf.Grow(int(off) - w.buf.Len())
	}
	n, err := w.buf.Write(p)
	return n, err
}

func TestS3Store_UploadAndDeletePublicWithContentType(t *testing.T) {
	store := setupS3Store(t)
	ctx := context.Background()
	key := "images/" + xid.New().String() + ".jpg"
	content := "fake jpeg data"

	err := store.UploadPublic(ctx, key, strings.NewReader(content))
	require.NoError(t, err)

	err = store.DeletePublic(ctx, key)
	require.NoError(t, err)
}

func TestS3Store_GetEndpoint(t *testing.T) {
	store := setupS3Store(t)
	endpoint := store.GetEndpoint()
	assert.Equal(t, "http://localhost:9000", endpoint)
}

func TestS3Store_GetBucketNames(t *testing.T) {
	store := setupS3Store(t)
	assert.Equal(t, "gama-public", store.GetPublicBucket())
	assert.Equal(t, "gama-private", store.GetPrivateBucket())
}

func TestS3Store_UploadDownloadRoundTrip(t *testing.T) {
	store := setupS3Store(t)
	ctx := context.Background()
	key := "roundtrip/" + xid.New().String() + ".json"
	original := `{"name": "test", "value": 42}`

	err := store.UploadPrivate(ctx, key, strings.NewReader(original))
	require.NoError(t, err)

	body, err := store.DownloadStreamPrivate(ctx, key)
	require.NoError(t, err)
	defer body.Close()

	downloaded, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, original, string(downloaded))

	store.DeletePrivate(ctx, key)
}

func TestS3Store_UploadReplaceExisting(t *testing.T) {
	store := setupS3Store(t)
	ctx := context.Background()
	key := "replace/" + xid.New().String() + ".txt"

	err := store.UploadPrivate(ctx, key, strings.NewReader("version 1"))
	require.NoError(t, err)

	err = store.UploadPrivate(ctx, key, strings.NewReader("version 2"))
	require.NoError(t, err)

	body, _ := store.DownloadStreamPrivate(ctx, key)
	data, _ := io.ReadAll(body)
	body.Close()
	assert.Equal(t, "version 2", string(data))

	store.DeletePrivate(ctx, key)
}

func TestS3Store_GeneratePresignedURL(t *testing.T) {
	store := setupS3Store(t)
	ctx := context.Background()
	key := "presigned/" + xid.New().String() + ".txt"

	err := store.UploadPrivate(ctx, key, strings.NewReader("presigned content"))
	require.NoError(t, err)

	url, err := store.GeneratePresignedURL(ctx, key, 5*time.Minute)
	require.NoError(t, err)
	assert.Contains(t, url, key)
	assert.Contains(t, url, "X-Amz-Signature")

	resp, err := httpGet(url)
	if err == nil {
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
	}

	store.DeletePrivate(ctx, key)
}

func httpGet(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Get(url)
}
