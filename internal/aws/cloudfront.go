package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

type CloudFrontClient struct {
	client *cloudfront.Client
}

func NewCloudFrontClient(ctx context.Context, profile string) (*CloudFrontClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &CloudFrontClient{
		client: cloudfront.NewFromConfig(cfg),
	}, nil
}

type CFDistributionInfo struct {
	ID         string
	Name       string
	Status     string
	Domain     string
	Comment    string
	Enabled    bool
	FirstAlias string
}

func (c *CloudFrontClient) ListDistributions(ctx context.Context) ([]CFDistributionInfo, error) {
	output, err := c.client.ListDistributions(ctx, &cloudfront.ListDistributionsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list distributions: %w", err)
	}

	if output.DistributionList == nil {
		return []CFDistributionInfo{}, nil
	}

	distros := make([]CFDistributionInfo, len(output.DistributionList.Items))
	for i, d := range output.DistributionList.Items {
		alias := ""
		if d.Aliases != nil && len(d.Aliases.Items) > 0 {
			alias = d.Aliases.Items[0]
		}
		distros[i] = CFDistributionInfo{
			ID:         aws.ToString(d.Id),
			Status:     aws.ToString(d.Status),
			Domain:     aws.ToString(d.DomainName),
			Comment:    aws.ToString(d.Comment),
			Enabled:    aws.ToBool(d.Enabled),
			FirstAlias: alias,
		}
	}

	return distros, nil
}

type CFOriginInfo struct {
	ID     string
	Domain string
	Path   string
}

type CFBehaviorInfo struct {
	PathPattern     string
	TargetOriginID  string
	ViewerProtocol  string
}

func (c *CloudFrontClient) GetDistributionDetails(ctx context.Context, id string) ([]CFOriginInfo, []CFBehaviorInfo, error) {
	output, err := c.client.GetDistribution(ctx, &cloudfront.GetDistributionInput{
		Id: aws.String(id),
	})
	if err != nil {
		return nil, nil, err
	}

	var origins []CFOriginInfo
	if output.Distribution.DistributionConfig.Origins != nil {
		for _, o := range output.Distribution.DistributionConfig.Origins.Items {
			origins = append(origins, CFOriginInfo{
				ID:     aws.ToString(o.Id),
				Domain: aws.ToString(o.DomainName),
				Path:   aws.ToString(o.OriginPath),
			})
		}
	}

	var behaviors []CFBehaviorInfo
	if output.Distribution.DistributionConfig.CacheBehaviors != nil {
		for _, b := range output.Distribution.DistributionConfig.CacheBehaviors.Items {
			behaviors = append(behaviors, CFBehaviorInfo{
				PathPattern:    aws.ToString(b.PathPattern),
				TargetOriginID: aws.ToString(b.TargetOriginId),
				ViewerProtocol: string(b.ViewerProtocolPolicy),
			})
		}
	}
	// Default behavior
	if output.Distribution.DistributionConfig.DefaultCacheBehavior != nil {
		b := output.Distribution.DistributionConfig.DefaultCacheBehavior
		behaviors = append(behaviors, CFBehaviorInfo{
			PathPattern:    "Default (*)",
			TargetOriginID: aws.ToString(b.TargetOriginId),
			ViewerProtocol: string(b.ViewerProtocolPolicy),
		})
	}

	return origins, behaviors, nil
}

type CFInvalidationInfo struct {
	ID          string
	Status      string
	CreateTime  string
}

func (c *CloudFrontClient) ListInvalidations(ctx context.Context, distributionID string) ([]CFInvalidationInfo, error) {
	output, err := c.client.ListInvalidations(ctx, &cloudfront.ListInvalidationsInput{
		DistributionId: aws.String(distributionID),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list invalidations: %w", err)
	}

	if output.InvalidationList == nil {
		return []CFInvalidationInfo{}, nil
	}

	invalidations := make([]CFInvalidationInfo, len(output.InvalidationList.Items))
	for i, v := range output.InvalidationList.Items {
		createTime := ""
		if v.CreateTime != nil {
			createTime = v.CreateTime.Format("2006-01-02 15:04:05")
		}
		invalidations[i] = CFInvalidationInfo{
			ID:         aws.ToString(v.Id),
			Status:     aws.ToString(v.Status),
			CreateTime: createTime,
		}
	}

	return invalidations, nil
}

type CFPolicyInfo struct {
	ID   string
	Name string
	Type string
}

func (c *CloudFrontClient) ListResponseHeadersPolicies(ctx context.Context) ([]CFPolicyInfo, error) {
	output, err := c.client.ListResponseHeadersPolicies(ctx, &cloudfront.ListResponseHeadersPoliciesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list response headers policies: %w", err)
	}

	if output.ResponseHeadersPolicyList == nil {
		return []CFPolicyInfo{}, nil
	}

	policies := make([]CFPolicyInfo, len(output.ResponseHeadersPolicyList.Items))
	for i, p := range output.ResponseHeadersPolicyList.Items {
		policies[i] = CFPolicyInfo{
			ID:   aws.ToString(p.ResponseHeadersPolicy.Id),
			Name: aws.ToString(p.ResponseHeadersPolicy.ResponseHeadersPolicyConfig.Name),
			Type: "ResponseHeaders",
		}
	}
	return policies, nil
}

type CFFunctionInfo struct {
	Name    string
	Status  string
	Runtime string
}

func (c *CloudFrontClient) ListFunctions(ctx context.Context) ([]CFFunctionInfo, error) {
	output, err := c.client.ListFunctions(ctx, &cloudfront.ListFunctionsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list functions: %w", err)
	}

	if output.FunctionList == nil {
		return []CFFunctionInfo{}, nil
	}

	fns := make([]CFFunctionInfo, len(output.FunctionList.Items))
	for i, f := range output.FunctionList.Items {
		fns[i] = CFFunctionInfo{
			Name:    aws.ToString(f.Name),
			Status:  aws.ToString(f.Status),
			Runtime: string(f.FunctionConfig.Runtime),
		}
	}

	return fns, nil
}
