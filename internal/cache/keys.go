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
	TTLDMSResources        = 10 * time.Minute // DMS resources
	TTLECSResources        = 10 * time.Minute // ECS resources
	TTLBillingResources    = 30 * time.Minute // Billing resources (longer TTL as it's less frequent)
	TTLSecurityHubResources = 10 * time.Minute // Security Hub resources
	TTLWAFResources        = 10 * time.Minute // WAF resources
	TTLECRResources        = 10 * time.Minute // ECR resources
	TTLEFSResources        = 10 * time.Minute // EFS resources
	TTLBackupResources      = 10 * time.Minute // AWS Backup resources
	TTLDynamoDBResources    = 10 * time.Minute // DynamoDB resources
	TTLTransferResources    = 10 * time.Minute // AWS Transfer resources
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

// DMSResources returns the cache key for DMS resources
func (kb *KeyBuilder) DMSResources(resourceType string) string {
	return fmt.Sprintf("%s:dms:%s", kb.profile, resourceType)
}

// ECSResources returns the cache key for ECS resources
func (kb *KeyBuilder) ECSResources(resourceType string) string {
	return fmt.Sprintf("%s:ecs:%s", kb.profile, resourceType)
}

// BillingResources returns the cache key for Billing resources
func (kb *KeyBuilder) BillingResources() string {
	return fmt.Sprintf("%s:billing", kb.profile)
}

// SecurityHubResources returns the cache key for Security Hub resources
func (kb *KeyBuilder) SecurityHubResources() string {
	return fmt.Sprintf("%s:securityhub", kb.profile)
}

// WAFResources returns the cache key for WAF resources
func (kb *KeyBuilder) WAFResources(resourceType string, scope string) string {
	return fmt.Sprintf("%s:waf:%s:%s", kb.profile, resourceType, scope)
}

// ECRResources returns the cache key for ECR resources
func (kb *KeyBuilder) ECRResources(resourceType string) string {
	return fmt.Sprintf("%s:ecr:%s", kb.profile, resourceType)
}

// ECRImages returns the cache key for ECR images in a repository
func (kb *KeyBuilder) ECRImages(repositoryName string) string {
	return fmt.Sprintf("%s:ecr:repository:%s:images", kb.profile, repositoryName)
}

// EFSResources returns the cache key for EFS resources
func (kb *KeyBuilder) EFSResources(resourceType string) string {
	return fmt.Sprintf("%s:efs:%s", kb.profile, resourceType)
}

// EFSMountTargets returns the cache key for EFS mount targets in a file system
func (kb *KeyBuilder) EFSMountTargets(fileSystemId string) string {
	return fmt.Sprintf("%s:efs:filesystem:%s:mount-targets", kb.profile, fileSystemId)
}

// BackupResources returns the cache key for AWS Backup resources
func (kb *KeyBuilder) BackupResources(resourceType string) string {
	return fmt.Sprintf("%s:backup:%s", kb.profile, resourceType)
}

// DynamoDBResources returns the cache key for DynamoDB resources
func (kb *KeyBuilder) DynamoDBResources(resourceType string) string {
	return fmt.Sprintf("%s:dynamodb:%s", kb.profile, resourceType)
}

// TransferResources returns the cache key for AWS Transfer resources
func (kb *KeyBuilder) TransferResources(resourceType string) string {
	return fmt.Sprintf("%s:transfer:%s", kb.profile, resourceType)
}

// TransferUsers returns the cache key for AWS Transfer users in a server
func (kb *KeyBuilder) TransferUsers(serverId string) string {
	return fmt.Sprintf("%s:transfer:server:%s:users", kb.profile, serverId)
}
