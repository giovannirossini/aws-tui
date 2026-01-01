package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSClient struct {
	client *sns.Client
}

func NewSNSClient(ctx context.Context, profile string) (*SNSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &SNSClient{
		client: sns.NewFromConfig(cfg),
	}, nil
}

type TopicInfo struct {
	ARN                string
	Name               string
	Type               string // Standard or FIFO
	SubscriptionsConfirmed string
	SubscriptionsPending   string
}

func (c *SNSClient) ListTopics(ctx context.Context) ([]TopicInfo, error) {
	var topicArns []string
	paginator := sns.NewListTopicsPaginator(c.client, &sns.ListTopicsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list topics: %w", err)
		}

		for _, t := range page.Topics {
			topicArns = append(topicArns, aws.ToString(t.TopicArn))
		}
	}

	// Fetch attributes in parallel
	var wg sync.WaitGroup
	resultChan := make(chan TopicInfo, len(topicArns))
	sem := make(chan struct{}, 10)

	for _, arn := range topicArns {
		wg.Add(1)
		go func(topicArn string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			name := topicArn[strings.LastIndex(topicArn, ":")+1:]
			topicType := "Standard"
			if strings.HasSuffix(name, ".fifo") {
				topicType = "FIFO"
			}

			attrOutput, err := c.client.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
				TopicArn: aws.String(topicArn),
			})

			var confirmed, pending string
			if err == nil {
				confirmed = attrOutput.Attributes["SubscriptionsConfirmed"]
				pending = attrOutput.Attributes["SubscriptionsPending"]
			}

			resultChan <- TopicInfo{
				ARN:                    topicArn,
				Name:                   name,
				Type:                   topicType,
				SubscriptionsConfirmed: confirmed,
				SubscriptionsPending:   pending,
			}
		}(arn)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var topics []TopicInfo
	for t := range resultChan {
		topics = append(topics, t)
	}

	return topics, nil
}
