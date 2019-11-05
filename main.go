package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
)

type BucketConfig struct {
	Region  string
	Bucket  string
	Prefix  string
	Profile string
}

type Config struct {
	Source      BucketConfig
	Destination BucketConfig
	Parallel    int
}

func main() {

	config := setup()
	sessSource := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(config.Source.Region),
		Credentials: credentials.NewSharedCredentials("", config.Source.Profile),
		MaxRetries:  aws.Int(10),
	}))

	sessDest := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(config.Destination.Region),
		Credentials: credentials.NewSharedCredentials("", config.Destination.Profile),
		MaxRetries:  aws.Int(10),
	}))

	ctx, cancelFunc := context.WithCancel(context.Background())
	handleInterrupt(cancelFunc)

	source := NewBucket(config.Source.Bucket, config.Source.Prefix, sessSource, ctx)
	dest := NewBucket(config.Destination.Bucket, config.Destination.Prefix, sessDest, ctx)
	j := job{
		source:      source,
		destination: dest,
		copyChan:    make(chan string, config.Parallel),
	}
	j.Start()
	log.Println("All done")
}
