package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
)

type MSKClient struct {
	client *kafka.Client
}

func NewMSKClient(ctx context.Context, profile string) (*MSKClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &MSKClient{
		client: kafka.NewFromConfig(cfg),
	}, nil
}

type ClusterInfo struct {
	ARN           string
	Name          string
	Status        string
	EngineVersion string
	NodeType      string
	Nodes         int32
}

func (c *MSKClient) ListClusters(ctx context.Context) ([]ClusterInfo, error) {
	output, err := c.client.ListClusters(ctx, &kafka.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list clusters: %w", err)
	}

	var clusters []ClusterInfo
	for _, cluster := range output.ClusterInfoList {
		nodes := int32(0)
		if cluster.BrokerNodeGroupInfo != nil {
			nodes = aws.ToInt32(cluster.NumberOfBrokerNodes)
		}

		clusters = append(clusters, ClusterInfo{
			ARN:           aws.ToString(cluster.ClusterArn),
			Name:          aws.ToString(cluster.ClusterName),
			Status:        string(cluster.State),
			EngineVersion: aws.ToString(cluster.CurrentVersion),
			NodeType:      "", // Not directly available in ListClusters without Describe
			Nodes:         nodes,
		})
	}

	return clusters, nil
}

func (c *MSKClient) ListClustersV2(ctx context.Context) ([]ClusterInfo, error) {
	output, err := c.client.ListClustersV2(ctx, &kafka.ListClustersV2Input{})
	if err != nil {
		return nil, fmt.Errorf("unable to list clusters v2: %w", err)
	}

	var clusters []ClusterInfo
	for _, cluster := range output.ClusterInfoList {
		nodes := int32(0)
		var version string
		var state string
		var name string
		var arn string

		if cluster.Provisioned != nil {
			nodes = aws.ToInt32(cluster.Provisioned.NumberOfBrokerNodes)
			version = aws.ToString(cluster.Provisioned.CurrentBrokerSoftwareInfo.KafkaVersion)
			state = string(cluster.State)
			name = aws.ToString(cluster.ClusterName)
			arn = aws.ToString(cluster.ClusterArn)
		} else if cluster.Serverless != nil {
			state = string(cluster.State)
			name = aws.ToString(cluster.ClusterName)
			arn = aws.ToString(cluster.ClusterArn)
		}

		clusters = append(clusters, ClusterInfo{
			ARN:           arn,
			Name:          name,
			Status:        state,
			EngineVersion: version,
			Nodes:         nodes,
		})
	}

	return clusters, nil
}
