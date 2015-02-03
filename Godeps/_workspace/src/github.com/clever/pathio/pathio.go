/*
Package pathio is a package that allows writing to and reading from different types of paths transparently.
It supports two types of paths:
 1. Local file paths
 2. S3 File Paths (s3://bucket/key)

Note that using s3 paths requires setting three environment variables
 1. AWS_SECRET_ACCESS_KEY
 2. AWS_ACCESS_KEY_ID
 3. AWS_REGION
*/
package pathio

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

// Reader returns an io.Reader for the specified path. The path can either be a local file path
// or an S3 path.
func Reader(path string) (io.Reader, error) {
	if strings.HasPrefix(path, "s3://") {
		return s3FileReader(path)
	}
	// Local file path
	return os.Open(path)
}

// Write writes a byte array to the specified path. The path can be either a local file path of an
// S3 path.
func Write(path string, input []byte) error {
	return WriteReader(path, bytes.NewReader(input), int64(len(input)))
}

// WriteReader writes all the data read from the specified io.Reader to the
// output path. The path can either a local file path or an S3 path.
func WriteReader(path string, input io.Reader, length int64) error {
	if strings.HasPrefix(path, "s3://") {
		return writeToS3(path, input, length)
	}
	return writeToLocalFile(path, input)

}

// s3FileReader converts an S3Path into an io.Reader
func s3FileReader(path string) (io.Reader, error) {
	bucket, key, err := getS3BucketAndKey(path)
	if err != nil {
		return nil, err
	}
	log.Printf("Getting from s3: %s", key)
	return bucket.GetReader(key)
}

func writeToLocalFile(path string, input io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, input)
	return err

}

func writeToS3(path string, input io.Reader, length int64) error {
	bucket, key, err := getS3BucketAndKey(path)
	if err != nil {
		return err
	}
	log.Printf("Writing to s3: %s", key)
	return bucket.PutReader(key, input, length, "text/plain", s3.Private)
}

// getS3BucketAndKey takes in a full s3path (s3://bucket/key) and returns a bucket,
// key, error tuple. It assumes that AWS environment variables are set.
func getS3BucketAndKey(path string) (*s3.Bucket, string, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Print("AWS environment variables not set")
		return nil, "", err
	}

	region, err := region(os.Getenv("AWS_REGION"))
	// This is a HACK, but the S3 library we use doesn't support redirections from Amazon, so when
	// we make a request to https://s3-us-west-1.amazonaws.com and Amazon returns a 301 redirecting
	// to https://s3.amazonaws.com the library blows up.
	region.S3Endpoint = "https://s3.amazonaws.com"
	if err != nil {
		return nil, "", err
	}
	s := s3.New(auth, region)
	bucketName, key, err := parseS3Path(path)
	if err != nil {
		return nil, "", err
	}
	return s.Bucket(bucketName), key, err
}

// parseS3path parses an S3 path (s3://bucket/key) and returns a bucket, key, error tuple
func parseS3Path(path string) (string, string, error) {
	// S3 path names are of the form s3://bucket/key
	stringsArray := strings.SplitN(path, "/", 4)
	if len(stringsArray) < 4 {
		return "", "", fmt.Errorf("Invalid s3 path %s", path)
	}
	bucketName := stringsArray[2]
	// Everything after the third slash is the key
	key := stringsArray[3]
	return bucketName, key, nil
}

// getRegion converts a region name into an aws.Region object
func region(regionString string) (aws.Region, error) {
	for name, region := range aws.Regions {
		if strings.ToLower(name) == strings.ToLower(regionString) {
			return region, nil
		}
	}
	return aws.Region{}, fmt.Errorf("Unknown region %s: ", regionString)
}
