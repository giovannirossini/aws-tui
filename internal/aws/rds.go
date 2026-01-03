package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

type RDSClient struct {
	client *rds.Client
}

func NewRDSClient(ctx context.Context, profile string) (*RDSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &RDSClient{
		client: rds.NewFromConfig(cfg),
	}, nil
}

type RDSInstanceInfo struct {
	ID       string
	Engine   string
	Status   string
	Class    string
	Endpoint string
	VpcID    string
}

func (c *RDSClient) ListInstances(ctx context.Context) ([]RDSInstanceInfo, error) {
	output, err := c.client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list RDS instances: %w", err)
	}

	instances := make([]RDSInstanceInfo, len(output.DBInstances))
	for i, d := range output.DBInstances {
		endpoint := ""
		if d.Endpoint != nil {
			endpoint = aws.ToString(d.Endpoint.Address)
		}
		vpcID := ""
		if d.DBSubnetGroup != nil {
			vpcID = aws.ToString(d.DBSubnetGroup.VpcId)
		}
		instances[i] = RDSInstanceInfo{
			ID:       aws.ToString(d.DBInstanceIdentifier),
			Engine:   aws.ToString(d.Engine),
			Status:   aws.ToString(d.DBInstanceStatus),
			Class:    aws.ToString(d.DBInstanceClass),
			Endpoint: endpoint,
			VpcID:    vpcID,
		}
	}

	return instances, nil
}

type RDSClusterInfo struct {
	ID       string
	Engine   string
	Status   string
	Endpoint string
	VpcID    string
}

func (c *RDSClient) ListClusters(ctx context.Context) ([]RDSClusterInfo, error) {
	output, err := c.client.DescribeDBClusters(ctx, &rds.DescribeDBClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list RDS clusters: %w", err)
	}

	clusters := make([]RDSClusterInfo, len(output.DBClusters))
	for i, d := range output.DBClusters {
		clusters[i] = RDSClusterInfo{
			ID:       aws.ToString(d.DBClusterIdentifier),
			Engine:   aws.ToString(d.Engine),
			Status:   aws.ToString(d.Status),
			Endpoint: aws.ToString(d.Endpoint),
			VpcID:    "", // VpcId is not directly in DBCluster, usually inferred from subnet group
		}
	}

	return clusters, nil
}

type RDSSnapshotInfo struct {
	ID         string
	InstanceID string
	Status     string
	Type       string
	CreateTime string
}

func (c *RDSClient) ListSnapshots(ctx context.Context) ([]RDSSnapshotInfo, error) {
	output, err := c.client.DescribeDBSnapshots(ctx, &rds.DescribeDBSnapshotsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list RDS snapshots: %w", err)
	}

	snapshots := make([]RDSSnapshotInfo, len(output.DBSnapshots))
	for i, d := range output.DBSnapshots {
		createTime := ""
		if d.SnapshotCreateTime != nil {
			createTime = d.SnapshotCreateTime.Format("2006-01-02 15:04:05")
		}
		snapshots[i] = RDSSnapshotInfo{
			ID:         aws.ToString(d.DBSnapshotIdentifier),
			InstanceID: aws.ToString(d.DBInstanceIdentifier),
			Status:     aws.ToString(d.Status),
			Type:       aws.ToString(d.SnapshotType),
			CreateTime: createTime,
		}
	}

	return snapshots, nil
}

type RDSSubnetGroupInfo struct {
	Name        string
	Description string
	VpcID       string
	Status      string
}

func (c *RDSClient) ListSubnetGroups(ctx context.Context) ([]RDSSubnetGroupInfo, error) {
	output, err := c.client.DescribeDBSubnetGroups(ctx, &rds.DescribeDBSubnetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list RDS subnet groups: %w", err)
	}

	groups := make([]RDSSubnetGroupInfo, len(output.DBSubnetGroups))
	for i, d := range output.DBSubnetGroups {
		groups[i] = RDSSubnetGroupInfo{
			Name:        aws.ToString(d.DBSubnetGroupName),
			Description: aws.ToString(d.DBSubnetGroupDescription),
			VpcID:       aws.ToString(d.VpcId),
			Status:      aws.ToString(d.SubnetGroupStatus),
		}
	}

	return groups, nil
}
