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
	TTLEC2Resources    = 10 * time.Minute // EC2 resources (Instances, SG, etc)
	TTLRDSResources    = 10 * time.Minute // RDS resources (Instances, Clusters, etc)
	TTLCWResources     = 5 * time.Minute  // CloudWatch resources
	TTLCFResources     = 10 * time.Minute // CloudFront resources
	TTLElastiCacheResources = 10 * time.Minute // ElastiCache resources
	TTLMSKResources        = 10 * time.Minute // MSK resources
	TTLSQSResources        = 10 * time.Minute // SQS resources
	TTLSMResources         = 10 * time.Minute // Secrets Manager resources
	TTLRoute53Resources    = 10 * time.Minute // Route 53 resources
	TTLACMResources        = 10 * time.Minute // ACM resources
	TTLSNSResources        = 10 * time.Minute // SNS resources
	TTLKMSResources        = 10 * time.Minute // KMS resources
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

// EC2Resources returns the cache key for EC2 resources
func (kb *KeyBuilder) EC2Resources(resourceType string) string {
	return fmt.Sprintf("%s:ec2:%s", kb.profile, resourceType)
}

// RDSResources returns the cache key for RDS resources
func (kb *KeyBuilder) RDSResources(resourceType string) string {
	return fmt.Sprintf("%s:rds:%s", kb.profile, resourceType)
}

// CWResources returns the cache key for CloudWatch resources
func (kb *KeyBuilder) CWResources(resourceType string) string {
	return fmt.Sprintf("%s:cw:%s", kb.profile, resourceType)
}

// CFResources returns the cache key for CloudFront resources
func (kb *KeyBuilder) CFResources(resourceType string) string {
	return fmt.Sprintf("%s:cf:%s", kb.profile, resourceType)
}

// ElastiCacheResources returns the cache key for ElastiCache resources
func (kb *KeyBuilder) ElastiCacheResources(resourceType string) string {
	return fmt.Sprintf("%s:elasticache:%s", kb.profile, resourceType)
}

// MSKResources returns the cache key for MSK resources
func (kb *KeyBuilder) MSKResources(resourceType string) string {
	return fmt.Sprintf("%s:msk:%s", kb.profile, resourceType)
}

// SQSResources returns the cache key for SQS resources
func (kb *KeyBuilder) SQSResources(resourceType string) string {
	return fmt.Sprintf("%s:sqs:%s", kb.profile, resourceType)
}

// SMResources returns the cache key for Secrets Manager resources
func (kb *KeyBuilder) SMResources(resourceType string) string {
	return fmt.Sprintf("%s:sm:%s", kb.profile, resourceType)
}

// Route53Resources returns the cache key for Route 53 resources
func (kb *KeyBuilder) Route53Resources(resourceType string) string {
	return fmt.Sprintf("%s:route53:%s", kb.profile, resourceType)
}

// ACMResources returns the cache key for ACM resources
func (kb *KeyBuilder) ACMResources(resourceType string) string {
	return fmt.Sprintf("%s:acm:%s", kb.profile, resourceType)
}

// SNSResources returns the cache key for SNS resources
func (kb *KeyBuilder) SNSResources(resourceType string) string {
	return fmt.Sprintf("%s:sns:%s", kb.profile, resourceType)
}

// KMSResources returns the cache key for KMS resources
func (kb *KeyBuilder) KMSResources(resourceType string) string {
	return fmt.Sprintf("%s:kms:%s", kb.profile, resourceType)
}
