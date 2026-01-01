package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	client *dynamodb.Client
}

func NewDynamoDBClient(ctx context.Context, profile string) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &DynamoDBClient{
		client: dynamodb.NewFromConfig(cfg),
	}, nil
}

type DynamoTableInfo struct {
	Name           string
	Status         string
	ItemCount      int64
	TableSize      int64
	CreationTime   time.Time
	PartitionKey   string
	SortKey        string
	BillingMode    string
}

func (c *DynamoDBClient) ListTables(ctx context.Context) ([]DynamoTableInfo, error) {
	var tables []DynamoTableInfo
	paginator := dynamodb.NewListTablesPaginator(c.client, &dynamodb.ListTablesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list tables: %w", err)
		}

		for _, tableName := range output.TableNames {
			desc, err := c.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				// Skip tables we can't describe
				continue
			}

			t := desc.Table
			info := DynamoTableInfo{
				Name:         aws.ToString(t.TableName),
				Status:       string(t.TableStatus),
				ItemCount:    aws.ToInt64(t.ItemCount),
				TableSize:    aws.ToInt64(t.TableSizeBytes),
				CreationTime: aws.ToTime(t.CreationDateTime),
			}

			for _, attr := range t.KeySchema {
				if attr.KeyType == "HASH" {
					info.PartitionKey = aws.ToString(attr.AttributeName)
				} else if attr.KeyType == "RANGE" {
					info.SortKey = aws.ToString(attr.AttributeName)
				}
			}

			if t.BillingModeSummary != nil {
				info.BillingMode = string(t.BillingModeSummary.BillingMode)
			} else {
				info.BillingMode = "PROVISIONED"
			}

			tables = append(tables, info)
		}
	}

	return tables, nil
}
