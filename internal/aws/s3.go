package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	client *s3.Client
}

func NewS3Client(ctx context.Context, profile string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &S3Client{
		client: s3.NewFromConfig(cfg),
	}, nil
}

type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

func (c *S3Client) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	output, err := c.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list buckets: %w", err)
	}

	buckets := make([]BucketInfo, len(output.Buckets))
	for i, b := range output.Buckets {
		buckets[i] = BucketInfo{
			Name:         aws.ToString(b.Name),
			CreationDate: aws.ToTime(b.CreationDate),
		}
	}

	return buckets, nil
}

type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	IsFolder     bool
}

func (c *S3Client) ListObjects(ctx context.Context, bucketName, prefix, delimiter string) ([]ObjectInfo, error) {
	output, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(delimiter),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list objects: %w", err)
	}

	var objects []ObjectInfo

	// Folders (CommonPrefixes)
	for _, cp := range output.CommonPrefixes {
		objects = append(objects, ObjectInfo{
			Key:      aws.ToString(cp.Prefix),
			IsFolder: true,
		})
	}

	// Objects
	for _, obj := range output.Contents {
		// Skip the folder itself if it's in the list
		if aws.ToString(obj.Key) == prefix {
			continue
		}
		objects = append(objects, ObjectInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
			IsFolder:     false,
		})
	}

	return objects, nil
}

func (c *S3Client) CreateBucket(ctx context.Context, name string, region string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	if region != "" && region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}
	_, err := c.client.CreateBucket(ctx, input)
	return err
}

func (c *S3Client) DeleteBucket(ctx context.Context, name string) error {
	_, err := c.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	return err
}

func (c *S3Client) CreateFolder(ctx context.Context, bucket, prefix string) error {
	// Ensure prefix ends with /
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix),
	})
	return err
}

func (c *S3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (c *S3Client) UploadFile(ctx context.Context, bucket, key, localPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}

func (c *S3Client) DownloadFile(ctx context.Context, bucket, key, localPath string) error {
	output, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer output.Body.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, output.Body)
	return err
}
