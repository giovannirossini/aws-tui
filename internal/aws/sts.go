package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type STSClient struct {
	client *sts.Client
	region string
}

func NewSTSClient(ctx context.Context, profile string) (*STSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &STSClient{
		client: sts.NewFromConfig(cfg),
		region: cfg.Region,
	}, nil
}

type IdentityInfo struct {
	Account string
	Arn     string
	UserId  string
	Alias   string
	Region  string
}

func (c *STSClient) GetCallerIdentity(ctx context.Context) (*IdentityInfo, error) {
	output, err := c.client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to get identity: %w", err)
	}

	return &IdentityInfo{
		Account: *output.Account,
		Arn:     *output.Arn,
		UserId:  *output.UserId,
		Region:  c.region,
	}, nil
}
