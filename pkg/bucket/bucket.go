// Package bucket manages the AWS S3 bucket uploads.
package bucket

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (

	// AccessKeyID is the env variable to check for the AWS access key id
	AccessKeyID = "AWS_ACCESS_KEY_ID"

	// AccessSecretKey is the env variable to check for the AWS secret key
	AccessSecretKey = "AWS_SECRET_ACCESS_KEY"

	// DefaultRegion is the AWS region to use for S3 bucket upload
	DefaultRegion = "eu-west-2"
)

var (
	errBucketName      = errors.New("bucket name is required")
	errBucketRegion    = errors.New("bucket region required")
	errAccessKeyID     = errors.New("no AWS_ACCESS_KEY_ID environment variable found")
	errAccessSecretKey = errors.New("no AWS_SECRET_ACCESS_KEY environment variable found")
)

// Bucket is used to pass AWS and S3
// information around Archer.
//
// TODO: this is used to interface with the
// S3 upload manager - I've used functional
// option setter so that we can add more
// fields here as needed when we have
// a clearer brief.
type Bucket struct {
	name            string
	region          string
	accessKeyID     string
	accessSecretKey string
}

// Option is a wrapper struct used to pass functional
// options to the Bucket constructor.
type Option func(bucket *Bucket) error

// SetName is an option setter for the New bucket constructor
// that sets the name field of a Bucket struct.
func SetName(name string) Option {
	return func(x *Bucket) error {
		if len(name) == 0 {
			return errBucketName
		}
		x.name = name
		return nil
	}
}

// SetRegion is an option setter for the New bucket constructor
// that sets the region field of a Bucket struct.
func SetRegion(region string) Option {
	return func(x *Bucket) error {
		if len(region) == 0 {
			return errBucketRegion
		}
		x.region = region
		return nil
	}
}

// New will construct a new bucket
// info struct.
func New(opts ...Option) (*Bucket, error) {
	b := &Bucket{
		region: DefaultRegion,
	}
	for _, opt := range opts {
		if err := opt(b); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// Check will check the bucket details and
// AWS authentication are provided.
func (b *Bucket) Check() error {

	// check for required info
	if len(b.name) == 0 {
		return errBucketName
	}
	if len(b.region) == 0 {
		return errBucketRegion
	}

	// get env variables
	b.accessKeyID = os.Getenv(AccessKeyID)
	b.accessSecretKey = os.Getenv(AccessSecretKey)
	if len(b.accessKeyID) == 0 {
		return errAccessKeyID
	}
	if len(b.accessSecretKey) == 0 {
		return errAccessSecretKey
	}
	return nil
}

// Upload will upload the contents of a reader
// to an S3 bucket using the provided key.
// It returns the upload location and any error.
func (b *Bucket) Upload(reader io.Reader, key string) (string, error) {

	// check the bucket details
	if err := b.Check(); err != nil {
		return "", err
	}

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(b.region)},
	)

	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("Failed to upload %v", err)
	}
	return result.Location, nil
}
