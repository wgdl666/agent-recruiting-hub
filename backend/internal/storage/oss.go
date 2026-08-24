package storage

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ResumeStore uploads resumes to Aliyun OSS. Falls back to local-only when not configured.
type ResumeStore struct {
	bucket *oss.Bucket
	prefix string
}

func FromEnv() *ResumeStore {
	endpoint := strings.TrimSpace(os.Getenv("HUB_OSS_ENDPOINT"))
	bucketName := strings.TrimSpace(os.Getenv("HUB_OSS_BUCKET"))
	ak := strings.TrimSpace(os.Getenv("HUB_OSS_ACCESS_KEY_ID"))
	sk := strings.TrimSpace(os.Getenv("HUB_OSS_ACCESS_KEY_SECRET"))
	prefix := strings.TrimSpace(os.Getenv("HUB_OSS_PREFIX"))
	if prefix == "" {
		prefix = "recruiting-hub/resumes"
	}
	prefix = strings.Trim(prefix, "/")
	if endpoint == "" || bucketName == "" || ak == "" || sk == "" {
		return &ResumeStore{prefix: prefix}
	}
	client, err := oss.New(endpoint, ak, sk)
	if err != nil {
		return &ResumeStore{prefix: prefix}
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return &ResumeStore{prefix: prefix}
	}
	return &ResumeStore{bucket: bucket, prefix: prefix}
}

func (r *ResumeStore) Enabled() bool {
	return r != nil && r.bucket != nil
}

func (r *ResumeStore) Prefix() string {
	if r == nil {
		return "recruiting-hub/resumes"
	}
	return r.prefix
}

func (r *ResumeStore) ObjectKey(filename string) string {
	name := strings.TrimPrefix(filepathBase(filename), "/")
	return r.Prefix() + "/" + name
}

func filepathBase(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func (r *ResumeStore) PutFile(localPath, key string) error {
	if !r.Enabled() {
		return fmt.Errorf("oss not configured")
	}
	return r.bucket.PutObjectFromFile(key, localPath)
}

func (r *ResumeStore) SignedGETURL(key string, expire time.Duration) (string, error) {
	if !r.Enabled() {
		return "", fmt.Errorf("oss not configured")
	}
	return r.bucket.SignURL(key, oss.HTTPGet, int64(expire.Seconds()))
}
