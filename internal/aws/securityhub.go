package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
)

type SecurityHubClient struct {
	client *securityhub.Client
}

func NewSecurityHubClient(ctx context.Context, profile string) (*SecurityHubClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &SecurityHubClient{
		client: securityhub.NewFromConfig(cfg),
	}, nil
}

type SecurityFinding struct {
	ID          string
	Title       string
	Severity    string
	Compliance  string
	ResourceID  string
	Region      string
	UpdatedAt   string
	Status      string
	Description string
}

func (c *SecurityHubClient) GetFindings(ctx context.Context) ([]SecurityFinding, error) {
	input := &securityhub.GetFindingsInput{
		Filters: &types.AwsSecurityFindingFilters{
			RecordState: []types.StringFilter{
				{
					Comparison: types.StringFilterComparisonEquals,
					Value:      aws.String("ACTIVE"),
				},
			},
			WorkflowStatus: []types.StringFilter{
				{
					Comparison: types.StringFilterComparisonNotEquals,
					Value:      aws.String("SUPPRESSED"),
				},
			},
		},
		MaxResults: aws.Int32(100),
	}

	output, err := c.client.GetFindings(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to get findings: %w", err)
	}

	var findings []SecurityFinding
	for _, f := range output.Findings {
		severity := "UNKNOWN"
		if f.Severity != nil {
			severity = string(f.Severity.Label)
		}

		compliance := "UNKNOWN"
		if f.Compliance != nil {
			compliance = string(f.Compliance.Status)
		}

		resourceID := "Unknown"
		if len(f.Resources) > 0 {
			resourceID = aws.ToString(f.Resources[0].Id)
		}

		findings = append(findings, SecurityFinding{
			ID:          aws.ToString(f.Id),
			Title:       aws.ToString(f.Title),
			Severity:    severity,
			Compliance:  compliance,
			ResourceID:  resourceID,
			Region:      aws.ToString(f.Region),
			UpdatedAt:   aws.ToString(f.UpdatedAt),
			Status:      string(f.Workflow.Status),
			Description: aws.ToString(f.Description),
		})
	}

	return findings, nil
}
