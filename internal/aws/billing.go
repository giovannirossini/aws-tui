package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type BillingClient struct {
	client *costexplorer.Client
}

func NewBillingClient(ctx context.Context, profile string) (*BillingClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &BillingClient{
		client: costexplorer.NewFromConfig(cfg),
	}, nil
}

type CostInfo struct {
	Service string
	Amount  string
	Unit    string
}

func (c *BillingClient) GetMonthlyCosts(ctx context.Context) ([]CostInfo, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	output, err := c.client.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to get cost and usage: %w", err)
	}

	var costs []CostInfo
	if len(output.ResultsByTime) > 0 {
		for _, group := range output.ResultsByTime[0].Groups {
			service := "Unknown"
			if len(group.Keys) > 0 {
				service = group.Keys[0]
			}

			amount := "0"
			unit := "USD"
			if cost, ok := group.Metrics["UnblendedCost"]; ok {
				amount = aws.ToString(cost.Amount)
				unit = aws.ToString(cost.Unit)
			}

			costs = append(costs, CostInfo{
				Service: service,
				Amount:  amount,
				Unit:    unit,
			})
		}
	}

	return costs, nil
}
