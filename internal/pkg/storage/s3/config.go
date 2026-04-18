package s3

import (
	"time"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	Bucket          string
	DefaultExpiry   time.Duration
}