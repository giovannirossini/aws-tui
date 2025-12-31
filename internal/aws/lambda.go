package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type LambdaClient struct {
	client *lambda.Client
}

func NewLambdaClient(ctx context.Context, profile string) (*LambdaClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &LambdaClient{
		client: lambda.NewFromConfig(cfg),
	}, nil
}

type FunctionInfo struct {
	Name         string
	Runtime      string
	Handler      string
	LastModified time.Time
	MemorySize   int32
	Timeout      int32
	Description  string
}

func (c *LambdaClient) ListFunctions(ctx context.Context) ([]FunctionInfo, error) {
	var functions []FunctionInfo
	paginator := lambda.NewListFunctionsPaginator(c.client, &lambda.ListFunctionsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list functions: %w", err)
		}

		for _, f := range output.Functions {
			lastModified, _ := time.Parse("2006-01-02T15:04:05.000-0700", aws.ToString(f.LastModified))
			functions = append(functions, FunctionInfo{
				Name:         aws.ToString(f.FunctionName),
				Runtime:      string(f.Runtime),
				Handler:      aws.ToString(f.Handler),
				LastModified: lastModified,
				MemorySize:   aws.ToInt32(f.MemorySize),
				Timeout:      aws.ToInt32(f.Timeout),
				Description:  aws.ToString(f.Description),
			})
		}
	}

	return functions, nil
}
