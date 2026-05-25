package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/xid"
)

type S3Store struct {
	client        *s3.Client
	uploader      *manager.Uploader
	downloader    *manager.Downloader
	presigner     *s3.PresignClient
	publicBucket  string
	privateBucket string
	publicURL     string
}

func NewS3Store(endpoint, region, accessKey, secretKey, publicBucket, privateBucket, publicURL string) (*S3Store, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &S3Store{
		client:        client,
		publicBucket:  publicBucket,
		privateBucket: privateBucket,
		publicURL:     publicURL,
		uploader: manager.NewUploader(client, func(u *manager.Uploader) {
			u.PartSize = 10 * 1024 * 1024
			u.Concurrency = 5
		}),
		downloader: manager.NewDownloader(client, func(d *manager.Downloader) {
			d.PartSize = 10 * 1024 * 1024
			d.Concurrency = 5
		}),
		presigner: s3.NewPresignClient(client),
	}, nil
}

func (s *S3Store) UploadPublic(ctx context.Context, key string, body io.Reader) error {
	contentType := detectContentTypeFromKey(key)

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.publicBucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to public bucket: %w", err)
	}
	return nil
}

func (s *S3Store) UploadPrivate(ctx context.Context, key string, body io.Reader) error {
	contentType := detectContentTypeFromKey(key)

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.privateBucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to private bucket: %w", err)
	}
	return nil
}

func (s *S3Store) DownloadPrivate(ctx context.Context, key string, writer io.WriterAt) error {
	_, err := s.downloader.Download(ctx, writer, &s3.GetObjectInput{
		Bucket: aws.String(s.privateBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download from private bucket: %w", err)
	}
	return nil
}

func (s *S3Store) DownloadStreamPrivate(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.privateBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from private bucket: %w", err)
	}
	return out.Body, nil
}

func (s *S3Store) DeletePublic(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.publicBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from public bucket: %w", err)
	}
	return nil
}

func (s *S3Store) DeletePrivate(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.privateBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from private bucket: %w", err)
	}
	return nil
}

func (s *S3Store) GeneratePresignedURL(ctx context.Context, key string, lifetime time.Duration) (string, error) {
	contentType := detectContentTypeFromKey(key)

	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:              aws.String(s.privateBucket),
		Key:                 aws.String(key),
		ResponseContentType: contentType,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = lifetime
	})
	if err != nil {
		return "", fmt.Errorf("failed to presign request: %w", err)
	}
	return req.URL, nil
}

func (s *S3Store) GetPublicBucket() string {
	return s.publicBucket
}

func (s *S3Store) GetPrivateBucket() string {
	return s.privateBucket
}

func (s *S3Store) GetPublicBaseURL() string {
	return s.publicURL
}

func (s *S3Store) GetPublicImageURL(key string) string {
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, key)
	}
	return key
}

func (s *S3Store) GetEndpoint() string {
	if s.client.Options().BaseEndpoint != nil {
		return *s.client.Options().BaseEndpoint
	}
	return ""
}

func GenerateUniqueKey(folder string, filename string) string {
	ext := filepath.Ext(filename)
	id := xid.New().String()
	return fmt.Sprintf("%s/%s%s", folder, id, ext)
}

func detectContentTypeFromKey(key string) *string {
	ext := strings.ToLower(filepath.Ext(key))
	if ext == "" {
		return nil
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".webp":
			contentType = "image/webp"
		case ".gif":
			contentType = "image/gif"
		case ".svg":
			contentType = "image/svg+xml"
		case ".pdf":
			contentType = "application/pdf"
		}
	}

	if contentType == "" {
		return nil
	}

	return aws.String(contentType)
}
