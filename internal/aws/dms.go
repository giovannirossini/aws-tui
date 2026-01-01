package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
)

type DMSClient struct {
	client *databasemigrationservice.Client
}

func NewDMSClient(ctx context.Context, profile string) (*DMSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &DMSClient{
		client: databasemigrationservice.NewFromConfig(cfg),
	}, nil
}

type ReplicationTaskInfo struct {
	ID        string
	ARN       string
	Status    string
	Type      string
	Source    string
	Target    string
	Instance  string
	FullLoadProgress int32
}

func (c *DMSClient) ListReplicationTasks(ctx context.Context) ([]ReplicationTaskInfo, error) {
	var tasks []ReplicationTaskInfo
	paginator := databasemigrationservice.NewDescribeReplicationTasksPaginator(c.client, &databasemigrationservice.DescribeReplicationTasksInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list replication tasks: %w", err)
		}

		for _, t := range page.ReplicationTasks {
			tasks = append(tasks, ReplicationTaskInfo{
				ID:        aws.ToString(t.ReplicationTaskIdentifier),
				ARN:       aws.ToString(t.ReplicationTaskArn),
				Status:    aws.ToString(t.Status),
				Type:      string(t.MigrationType),
				Source:    aws.ToString(t.SourceEndpointArn),
				Target:    aws.ToString(t.TargetEndpointArn),
				Instance:  aws.ToString(t.ReplicationInstanceArn),
				FullLoadProgress: t.ReplicationTaskStats.FullLoadProgressPercent,
			})
		}
	}

	return tasks, nil
}

type DMSEndpointInfo struct {
	ID       string
	Type     string
	Engine   string
	Server   string
	Port     int32
	Status   string
}

func (c *DMSClient) ListEndpoints(ctx context.Context) ([]DMSEndpointInfo, error) {
	var endpoints []DMSEndpointInfo
	paginator := databasemigrationservice.NewDescribeEndpointsPaginator(c.client, &databasemigrationservice.DescribeEndpointsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list endpoints: %w", err)
		}

		for _, e := range page.Endpoints {
			endpoints = append(endpoints, DMSEndpointInfo{
				ID:     aws.ToString(e.EndpointIdentifier),
				Type:   string(e.EndpointType),
				Engine: aws.ToString(e.EngineName),
				Server: aws.ToString(e.ServerName),
				Port:   aws.ToInt32(e.Port),
				Status: aws.ToString(e.Status),
			})
		}
	}

	return endpoints, nil
}

type ReplicationInstanceInfo struct {
	ID               string
	Class            string
	Status           string
	EngineVersion    string
	PubliclyAccessible bool
	AZ               string
}

func (c *DMSClient) ListReplicationInstances(ctx context.Context) ([]ReplicationInstanceInfo, error) {
	var instances []ReplicationInstanceInfo
	paginator := databasemigrationservice.NewDescribeReplicationInstancesPaginator(c.client, &databasemigrationservice.DescribeReplicationInstancesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list replication instances: %w", err)
		}

		for _, i := range page.ReplicationInstances {
			instances = append(instances, ReplicationInstanceInfo{
				ID:               aws.ToString(i.ReplicationInstanceIdentifier),
				Class:            aws.ToString(i.ReplicationInstanceClass),
				Status:           aws.ToString(i.ReplicationInstanceStatus),
				EngineVersion:    aws.ToString(i.EngineVersion),
				PubliclyAccessible: i.PubliclyAccessible,
				AZ:               aws.ToString(i.AvailabilityZone),
			})
		}
	}

	return instances, nil
}

func (c *DMSClient) StartReplicationTask(ctx context.Context, taskArn string, startType types.StartReplicationTaskTypeValue) error {
	_, err := c.client.StartReplicationTask(ctx, &databasemigrationservice.StartReplicationTaskInput{
		ReplicationTaskArn:       aws.String(taskArn),
		StartReplicationTaskType: startType,
	})
	return err
}

func (c *DMSClient) StopReplicationTask(ctx context.Context, taskArn string) error {
	_, err := c.client.StopReplicationTask(ctx, &databasemigrationservice.StopReplicationTaskInput{
		ReplicationTaskArn: aws.String(taskArn),
	})
	return err
}
