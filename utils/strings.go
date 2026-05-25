package utils

import (
	"crypto/rand"
	"path/filepath"
	"strings"
)

func SplitClean(s string, d string) []string {
	raw := strings.Split(s, d)

	out := make([]string, 0, len(raw))

	for _, v := range raw {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}

	return out
}

func JoinClean(items []string, d string) string {
	out := make([]string, 0, len(items))

	for _, v := range items {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}

	return strings.Join(out, d)
}

func GetContentTypeFromPath(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func IsDevEnv(mode string) bool {
	return EqualFoldASCII(mode, "dev")
}

func IsSubsetSlice(slice1, slice2 []string) bool {
	if len(slice1) < len(slice2) {
		return false
	}

	elements := make(map[string]struct{})

	for _, item := range slice2 {
		elements[item] = struct{}{}
	}

	for _, item := range slice1 {
		if _, found := elements[item]; !found {
			return false
		}
	}

	return true
}

func EqualFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range len(a) {
		ca := a[i]
		cb := b[i]

		if ca != cb {
			if ca >= 65 && ca <= 90 {
				ca += 32
			}

			if cb >= 65 && cb <= 90 {
				cb += 32
			}

			if ca != cb {
				return false
			}
		}
	}

	return true
}

func GenerateRandomNumber(length int) (string, error) {
	const otpChars = "0123456789"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%len(otpChars)]
	}

	return string(buffer), nil
}
