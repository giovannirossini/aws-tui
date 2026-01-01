package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type CloudWatchClient struct {
	client *cloudwatchlogs.Client
}

func NewCloudWatchClient(ctx context.Context, profile string) (*CloudWatchClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &CloudWatchClient{
		client: cloudwatchlogs.NewFromConfig(cfg),
	}, nil
}

type LogGroupInfo struct {
	Name            string
	RetentionDays   int32
	StoredBytes     int64
	CreationTime    string
	Arn             string
}

func (c *CloudWatchClient) ListLogGroups(ctx context.Context) ([]LogGroupInfo, error) {
	output, err := c.client.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list log groups: %w", err)
	}

	groups := make([]LogGroupInfo, len(output.LogGroups))
	for i, g := range output.LogGroups {
		creationTime := ""
		if g.CreationTime != nil {
			creationTime = time.Unix(*g.CreationTime/1000, 0).Format("2006-01-02 15:04:05")
		}
		groups[i] = LogGroupInfo{
			Name:          aws.ToString(g.LogGroupName),
			RetentionDays: aws.ToInt32(g.RetentionInDays),
			StoredBytes:   aws.ToInt64(g.StoredBytes),
			CreationTime:  creationTime,
			Arn:           aws.ToString(g.Arn),
		}
	}

	return groups, nil
}

type LogStreamInfo struct {
	Name                 string
	LastEventTimestamp   string
	CreationTime         string
	Arn                  string
}

func (c *CloudWatchClient) ListLogStreams(ctx context.Context, logGroupName string) ([]LogStreamInfo, error) {
	output, err := c.client.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName),
		OrderBy:      "LastEventTime",
		Descending:   aws.Bool(true),
		Limit:        aws.Int32(50),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list log streams: %w", err)
	}

	streams := make([]LogStreamInfo, len(output.LogStreams))
	for i, s := range output.LogStreams {
		lastEvent := ""
		if s.LastEventTimestamp != nil {
			lastEvent = time.Unix(*s.LastEventTimestamp/1000, 0).Format("2006-01-02 15:04:05")
		}
		creationTime := ""
		if s.CreationTime != nil {
			creationTime = time.Unix(*s.CreationTime/1000, 0).Format("2006-01-02 15:04:05")
		}
		streams[i] = LogStreamInfo{
			Name:               aws.ToString(s.LogStreamName),
			LastEventTimestamp: lastEvent,
			CreationTime:       creationTime,
			Arn:                aws.ToString(s.Arn),
		}
	}

	return streams, nil
}

type LogEventInfo struct {
	Timestamp int64
	Message   string
}

func (c *CloudWatchClient) GetLogEvents(ctx context.Context, logGroupName, logStreamName string) ([]LogEventInfo, error) {
	output, err := c.client.GetLogEvents(ctx, &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		Limit:         aws.Int32(100),
		StartFromHead: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get log events: %w", err)
	}

	events := make([]LogEventInfo, len(output.Events))
	for i, e := range output.Events {
		events[i] = LogEventInfo{
			Timestamp: aws.ToInt64(e.Timestamp),
			Message:   aws.ToString(e.Message),
		}
	}

	return events, nil
}
