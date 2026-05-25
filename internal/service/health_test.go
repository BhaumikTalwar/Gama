package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeStatus_NilComponents(t *testing.T) {
	response := &HealthResponse{
		Status:    "initial",
		Components: nil,
	}

	status := response.ComputeStatus()
	assert.Equal(t, "initial", status)
}

func TestComputeStatus_AllHealthy(t *testing.T) {
	response := &HealthResponse{
		Status:    "healthy",
		Components: map[string]ComponentHealth{
			"database": {Status: "healthy"},
			"redis":    {Status: "healthy"},
		},
	}

	status := response.ComputeStatus()
	assert.Equal(t, "healthy", status)
}

func TestComputeStatus_OneUnhealthy(t *testing.T) {
	response := &HealthResponse{
		Status:    "healthy",
		Components: map[string]ComponentHealth{
			"database": {Status: "healthy"},
			"redis":    {Status: "unhealthy", Error: "connection failed"},
		},
	}

	status := response.ComputeStatus()
	assert.Equal(t, "unhealthy", status)
}

func TestComputeStatus_OneDegraded(t *testing.T) {
	response := &HealthResponse{
		Status:    "healthy",
		Components: map[string]ComponentHealth{
			"database": {Status: "healthy"},
			"s3":       {Status: "degraded"},
		},
	}

	status := response.ComputeStatus()
	assert.Equal(t, "degraded", status)
}

func TestComputeStatus_UnhealthyTakesPrecedence(t *testing.T) {
	response := &HealthResponse{
		Status:    "healthy",
		Components: map[string]ComponentHealth{
			"database": {Status: "degraded"},
			"redis":    {Status: "unhealthy"},
			"s3":       {Status: "healthy"},
		},
	}

	status := response.ComputeStatus()
	assert.Equal(t, "unhealthy", status)
}

func TestComputeStatus_EmptyComponents(t *testing.T) {
	response := &HealthResponse{
		Status:     "healthy",
		Components: map[string]ComponentHealth{},
	}

	status := response.ComputeStatus()
	assert.Equal(t, "healthy", status)
}

func TestComputeStatus_MixedHealthyAndDegraded(t *testing.T) {
	response := &HealthResponse{
		Status:    "healthy",
		Components: map[string]ComponentHealth{
			"database": {Status: "healthy"},
			"redis":    {Status: "healthy"},
			"s3":       {Status: "degraded", Error: "S3 connection failed"},
		},
	}

	status := response.ComputeStatus()
	assert.Equal(t, "degraded", status)
}

func TestLiveness(t *testing.T) {
	hc := &HealthChecker{}
	health := hc.Liveness()

	assert.Equal(t, "alive", health.Status)
	assert.Empty(t, health.Latency)
	assert.Empty(t, health.Error)
}

func TestGenerateUniqueKey(t *testing.T) {
	key := GenerateUniqueKey("avatars", "photo.png")
	assert.Contains(t, key, "avatars/")
	assert.Contains(t, key, ".png")
	assert.NotEqual(t, key, "avatars/.png")
}

func TestGenerateUniqueKey_DifferentFiles(t *testing.T) {
	key1 := GenerateUniqueKey("folder", "file1.jpg")
	key2 := GenerateUniqueKey("folder", "file2.jpg")

	assert.NotEqual(t, key1, key2)
}

func TestGenerateUniqueKey_NoExtension(t *testing.T) {
	key := GenerateUniqueKey("files", "document")
	assert.Contains(t, key, "files/")
	assert.NotContains(t, key, "document")
	assert.NotContains(t, key, ".")
}

func TestGenerateUniqueKey_UniquePerCall(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key := GenerateUniqueKey("uploads", "file.jpg")
		keys[key] = true
	}
	assert.Greater(t, len(keys), 1)
}

func TestDetectContentTypeFromKey_JPEG(t *testing.T) {
	result := detectContentTypeFromKey("photo.jpg")
	assert.NotNil(t, result)
	assert.Equal(t, "image/jpeg", *result)
}

func TestDetectContentTypeFromKey_JPEGExtension(t *testing.T) {
	result := detectContentTypeFromKey("photo.jpeg")
	assert.NotNil(t, result)
	assert.Equal(t, "image/jpeg", *result)
}

func TestDetectContentTypeFromKey_PNG(t *testing.T) {
	result := detectContentTypeFromKey("image.png")
	assert.NotNil(t, result)
	assert.Equal(t, "image/png", *result)
}

func TestDetectContentTypeFromKey_GIF(t *testing.T) {
	result := detectContentTypeFromKey("image.gif")
	assert.NotNil(t, result)
	assert.Equal(t, "image/gif", *result)
}

func TestDetectContentTypeFromKey_WEBP(t *testing.T) {
	result := detectContentTypeFromKey("image.webp")
	assert.NotNil(t, result)
	assert.Equal(t, "image/webp", *result)
}

func TestDetectContentTypeFromKey_SVG(t *testing.T) {
	result := detectContentTypeFromKey("image.svg")
	assert.NotNil(t, result)
	assert.Equal(t, "image/svg+xml", *result)
}

func TestDetectContentTypeFromKey_PDF(t *testing.T) {
	result := detectContentTypeFromKey("doc.pdf")
	assert.NotNil(t, result)
	assert.Equal(t, "application/pdf", *result)
}

func TestDetectContentTypeFromKey_NoExtension(t *testing.T) {
	result := detectContentTypeFromKey("file")
	assert.Nil(t, result)
}

func TestDetectContentTypeFromKey_UnknownExtension(t *testing.T) {
	result := detectContentTypeFromKey("file.xyz")
	assert.Nil(t, result)
}

func TestDetectContentTypeFromKey_CaseInsensitive(t *testing.T) {
	result := detectContentTypeFromKey("file.JPG")
	assert.NotNil(t, result)
	assert.Equal(t, "image/jpeg", *result)
}

func TestS3Store_GetPublicBucket(t *testing.T) {
	store := &S3Store{publicBucket: "test-public"}
	assert.Equal(t, "test-public", store.GetPublicBucket())
}

func TestS3Store_GetPrivateBucket(t *testing.T) {
	store := &S3Store{privateBucket: "test-private"}
	assert.Equal(t, "test-private", store.GetPrivateBucket())
}

func TestS3Store_GetPublicBaseURL(t *testing.T) {
	store := &S3Store{publicURL: "https://cdn.example.com"}
	assert.Equal(t, "https://cdn.example.com", store.GetPublicBaseURL())
}

func TestS3Store_GetPublicImageURL_WithBaseURL(t *testing.T) {
	store := &S3Store{publicURL: "https://cdn.example.com"}
	url := store.GetPublicImageURL("avatars/user1.png")
	assert.Equal(t, "https://cdn.example.com/avatars/user1.png", url)
}

func TestS3Store_GetPublicImageURL_WithoutBaseURL(t *testing.T) {
	store := &S3Store{publicURL: ""}
	url := store.GetPublicImageURL("avatars/user1.png")
	assert.Equal(t, "avatars/user1.png", url)
}