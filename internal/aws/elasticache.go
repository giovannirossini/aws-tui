package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

type ElastiCacheClient struct {
	client *elasticache.Client
}

func NewElastiCacheClient(ctx context.Context, profile string) (*ElastiCacheClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &ElastiCacheClient{
		client: elasticache.NewFromConfig(cfg),
	}, nil
}

type ReplicationGroupInfo struct {
	ID          string
	Status      string
	Engine      string
	CacheNodeType string
	Nodes       int32
	Description string
}

func (c *ElastiCacheClient) ListReplicationGroups(ctx context.Context) ([]ReplicationGroupInfo, error) {
	output, err := c.client.DescribeReplicationGroups(ctx, &elasticache.DescribeReplicationGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list replication groups: %w", err)
	}

	var groups []ReplicationGroupInfo
	for _, rg := range output.ReplicationGroups {
		groups = append(groups, ReplicationGroupInfo{
			ID:            aws.ToString(rg.ReplicationGroupId),
			Status:        aws.ToString(rg.Status),
			Engine:        aws.ToString(rg.Engine),
			CacheNodeType: aws.ToString(rg.CacheNodeType),
			Nodes:         int32(len(rg.NodeGroups)),
			Description:   aws.ToString(rg.Description),
		})
	}

	return groups, nil
}

type CacheClusterInfo struct {
	ID            string
	Status        string
	Engine        string
	EngineVersion string
	CacheNodeType string
	Nodes         int32
	AZ            string
}

func (c *ElastiCacheClient) ListCacheClusters(ctx context.Context) ([]CacheClusterInfo, error) {
	output, err := c.client.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list cache clusters: %w", err)
	}

	var clusters []CacheClusterInfo
	for _, cc := range output.CacheClusters {
		clusters = append(clusters, CacheClusterInfo{
			ID:            aws.ToString(cc.CacheClusterId),
			Status:        aws.ToString(cc.CacheClusterStatus),
			Engine:        aws.ToString(cc.Engine),
			EngineVersion: aws.ToString(cc.EngineVersion),
			CacheNodeType: aws.ToString(cc.CacheNodeType),
			Nodes:         aws.ToInt32(cc.NumCacheNodes),
			AZ:            aws.ToString(cc.PreferredAvailabilityZone),
		})
	}

	return clusters, nil
}
