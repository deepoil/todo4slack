package config

import (
	"os"
)

type ConfigList struct {
	ApiToken     string
	ApiSecret    string
	AwsAccessKey string
	AwsSecretKey string
	BucketName   string
}

var Config ConfigList

func init() {
	Config = ConfigList{
		ApiToken:     os.Getenv("SLACK_API_TOKEN"),
		ApiSecret:    os.Getenv("SLACK_API_SECRET"),
		AwsAccessKey: os.Getenv("AWS_ACCESS_KEY"),
		AwsSecretKey: os.Getenv("AWS_SECRET_KEY"),
		BucketName:   os.Getenv("S3_BUCKET_NAME"),
	}
}
