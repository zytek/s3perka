package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
)

func main() {

	config := readConfig()
	// Initialize source bucket AWS Session
	sourceConfig := &aws.Config{
		Region:      aws.String(config.Source.Region),
		MaxRetries:  aws.Int(10),
	}
	if config.Source.Profile != "" {
		sourceConfig.Credentials = credentials.NewSharedCredentials("", config.Source.Profile)
	}

	sessSource := session.Must(session.NewSession(sourceConfig))

    // Initialize destination bucket AWS Session
	destConfig := &aws.Config{
		Region:      aws.String(config.Destination.Region),
		MaxRetries:  aws.Int(10),
	}

	if config.Destination.Profile != "" {
		destConfig.Credentials = credentials.NewSharedCredentials("", config.Destination.Profile)
	}
  	sessDest := session.Must(session.NewSession(destConfig))

	// Setup context and interrupt (ctrl+c) handling
  	ctx, cancelFunc := context.WithCancel(context.Background())
	handleInterrupt(cancelFunc)

  	// Initialize main copy job
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
