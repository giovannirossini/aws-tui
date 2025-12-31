package cache

import (
	"fmt"
	"time"
)

// Default TTL values for different AWS resource types
const (
	TTLIdentity       = 1 * time.Hour    // STS GetCallerIdentity + Account Alias
	TTLIAMUsers       = 5 * time.Minute  // IAM ListUsers
	TTLIAMUserDetails = 2 * time.Minute  // IAM User detailed info
	TTLIAMAccessKeys  = 2 * time.Minute  // IAM Access Keys
	TTLS3Buckets      = 10 * time.Minute // S3 ListBuckets
	TTLLongS3Objects  = 2 * time.Minute  // S3 ListObjects (large buckets)
	TTLShortS3Objects = 30 * time.Second // S3 ListObjects (small/active buckets)
	TTLVPCResources   = 10 * time.Minute // VPC resources (VPC, Subnets, etc)
	TTLLambdaFunctions = 5 * time.Minute  // Lambda ListFunctions
)

// KeyBuilder provides methods to build cache keys
type KeyBuilder struct {
	profile string
}

// NewKeyBuilder creates a new key builder for the given profile
func NewKeyBuilder(profile string) *KeyBuilder {
	return &KeyBuilder{profile: profile}
}

// Identity returns the cache key for identity info
func (kb *KeyBuilder) Identity() string {
	return fmt.Sprintf("%s:identity", kb.profile)
}

// IAMUsers returns the cache key for IAM users list
func (kb *KeyBuilder) IAMUsers() string {
	return fmt.Sprintf("%s:iam:users", kb.profile)
}

// IAMUserDetails returns the cache key for a specific IAM user
func (kb *KeyBuilder) IAMUserDetails(userName string) string {
	return fmt.Sprintf("%s:iam:user:%s", kb.profile, userName)
}

// S3Buckets returns the cache key for S3 buckets list
func (kb *KeyBuilder) S3Buckets() string {
	return fmt.Sprintf("%s:s3:buckets", kb.profile)
}

// S3Objects returns the cache key for S3 objects in a bucket/prefix
func (kb *KeyBuilder) S3Objects(bucket, prefix string) string {
	return fmt.Sprintf("%s:s3:bucket:%s:prefix:%s", kb.profile, bucket, prefix)
}

// ProfilePrefix returns the prefix for all cache keys for this profile
func (kb *KeyBuilder) ProfilePrefix() string {
	return kb.profile + ":"
}

// IAMPrefix returns the prefix for all IAM-related cache keys
func (kb *KeyBuilder) IAMPrefix() string {
	return fmt.Sprintf("%s:iam:", kb.profile)
}

// S3Prefix returns the prefix for all S3-related cache keys
func (kb *KeyBuilder) S3Prefix() string {
	return fmt.Sprintf("%s:s3:", kb.profile)
}

// S3BucketPrefix returns the prefix for all objects in a specific bucket
func (kb *KeyBuilder) S3BucketPrefix(bucket string) string {
	return fmt.Sprintf("%s:s3:bucket:%s:", kb.profile, bucket)
}

// VPCResources returns the cache key for VPC resources
func (kb *KeyBuilder) VPCResources(resourceType string) string {
	return fmt.Sprintf("%s:vpc:%s", kb.profile, resourceType)
}

// LambdaFunctions returns the cache key for Lambda functions list
func (kb *KeyBuilder) LambdaFunctions() string {
	return fmt.Sprintf("%s:lambda:functions", kb.profile)
}
