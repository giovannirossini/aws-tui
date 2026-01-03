package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

type WAFClient struct {
	client *wafv2.Client
}

func NewWAFClient(ctx context.Context, profile string, region string) (*WAFClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &WAFClient{
		client: wafv2.NewFromConfig(cfg),
	}, nil
}

type WebACLInfo struct {
	Name        string
	ID          string
	ARN         string
	Description string
}

func (c *WAFClient) ListWebACLs(ctx context.Context, scope types.Scope) ([]WebACLInfo, error) {
	input := &wafv2.ListWebACLsInput{
		Scope: scope,
		Limit: aws.Int32(100),
	}

	output, err := c.client.ListWebACLs(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to list web ACLs: %w", err)
	}

	var webACLs []WebACLInfo
	for _, acl := range output.WebACLs {
		webACLs = append(webACLs, WebACLInfo{
			Name:        aws.ToString(acl.Name),
			ID:          aws.ToString(acl.Id),
			ARN:         aws.ToString(acl.ARN),
			Description: aws.ToString(acl.Description),
		})
	}

	return webACLs, nil
}

type IPSetInfo struct {
	Name        string
	ID          string
	ARN         string
	Description string
}

func (c *WAFClient) ListIPSets(ctx context.Context, scope types.Scope) ([]IPSetInfo, error) {
	input := &wafv2.ListIPSetsInput{
		Scope: scope,
		Limit: aws.Int32(100),
	}

	output, err := c.client.ListIPSets(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to list IP sets: %w", err)
	}

	var ipSets []IPSetInfo
	for _, ipSet := range output.IPSets {
		ipSets = append(ipSets, IPSetInfo{
			Name:        aws.ToString(ipSet.Name),
			ID:          aws.ToString(ipSet.Id),
			ARN:         aws.ToString(ipSet.ARN),
			Description: aws.ToString(ipSet.Description),
		})
	}

	return ipSets, nil
}
