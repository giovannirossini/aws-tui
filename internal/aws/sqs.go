package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSClient struct {
	client *sqs.Client
}

func NewSQSClient(ctx context.Context, profile string) (*SQSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &SQSClient{
		client: sqs.NewFromConfig(cfg),
	}, nil
}

type QueueInfo struct {
	URL                string
	Name               string
	Type               string // Standard or FIFO
	MessagesAvailable  string
	MessagesDelayed    string
	MessagesNotVisible string
	VisibilityTimeout  string
	CreatedTimestamp   string
}

func (c *SQSClient) ListQueues(ctx context.Context) ([]QueueInfo, error) {
	output, err := c.client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list queues: %w", err)
	}

	var queues []QueueInfo
	for _, url := range output.QueueUrls {
		name := url[strings.LastIndex(url, "/")+1:]
		qType := "Standard"
		if strings.HasSuffix(name, ".fifo") {
			qType = "FIFO"
		}

		// Fetch attributes for each queue
		attrOutput, err := c.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(url),
			AttributeNames: []types.QueueAttributeName{
				types.QueueAttributeNameApproximateNumberOfMessages,
				types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
				types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
				types.QueueAttributeNameVisibilityTimeout,
				types.QueueAttributeNameCreatedTimestamp,
			},
		})

		var avail, delayed, notVisible, timeout, created string
		if err == nil {
			avail = attrOutput.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]
			delayed = attrOutput.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesDelayed)]
			notVisible = attrOutput.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesNotVisible)]
			timeout = attrOutput.Attributes[string(types.QueueAttributeNameVisibilityTimeout)]
			created = attrOutput.Attributes[string(types.QueueAttributeNameCreatedTimestamp)]
		}

		queues = append(queues, QueueInfo{
			URL:                url,
			Name:               name,
			Type:               qType,
			MessagesAvailable:  avail,
			MessagesDelayed:    delayed,
			MessagesNotVisible: notVisible,
			VisibilityTimeout:  timeout,
			CreatedTimestamp:   created,
		})
	}

	return queues, nil
}
