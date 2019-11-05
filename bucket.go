package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"strings"
)

type Bucket struct {
	Name       string
	Prefix     string
	Region     string
	Objects    map[string]*s3.Object
	TotalSize  int64
	TotalCount int
	session    *session.Session
	Context    context.Context
}

func (b *Bucket) processPage(p *s3.ListObjectsV2Output, last bool) (shouldContinue bool) {
	for _, object := range p.Contents {
		b.Objects[*object.Key] = object
		b.TotalSize += *object.Size
	}
	return true
}

func (b *Bucket) DownloadObject(key string, w io.WriterAt) (n int64, err error) {
	downloader := s3manager.NewDownloader(b.session)
	n, err = downloader.Download(w,
		&s3.GetObjectInput{
			Bucket: aws.String(b.Name),
			Key:    aws.String(key),
		})
	return
}

func (b *Bucket) UploadObject(key string, r io.Reader) (err error) {
	uploader := s3manager.NewUploader(b.session)
	_, err = uploader.Upload(
		&s3manager.UploadInput{
			Bucket: aws.String(b.Name),
			Key:    aws.String(key),
			Body:   r,
			ACL:    aws.String("private"),
		})
	return
}
func (b *Bucket) CollectKeys() (err error) {
	srv := s3.New(b.session)
	input := s3.ListObjectsV2Input{
		Bucket: aws.String(b.Name),
		Prefix: aws.String(b.Prefix),
	}
	err = srv.ListObjectsV2PagesWithContext(b.Context, &input, b.processPage)
	b.TotalCount = len(b.Objects)
	return err
}

func NewBucket(name string, prefix string, session *session.Session, ctx context.Context) *Bucket {
	b := &Bucket{
		Name:    name,
		Prefix:  strings.TrimPrefix(prefix, "/"),
		Region:  aws.StringValue(session.Config.Region),
		Objects: make(map[string]*s3.Object),
		session: session,
		Context: ctx,
	}
	return b
}
