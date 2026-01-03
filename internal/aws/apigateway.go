package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
)

type APIGatewayClient struct {
	client   *apigateway.Client
	clientV2 *apigatewayv2.Client
}

func NewAPIGatewayClient(ctx context.Context, profile string) (*APIGatewayClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &APIGatewayClient{
		client:   apigateway.NewFromConfig(cfg),
		clientV2: apigatewayv2.NewFromConfig(cfg),
	}, nil
}

type RestAPIInfo struct {
	ID           string
	Name         string
	Description  string
	CreatedDate  time.Time
	Version      string
	EndpointType string
}

type HTTPAPIInfo struct {
	APIID        string
	Name         string
	ProtocolType string
	CreatedDate  time.Time
	APIEndpoint  string
	Description  string
}

func (c *APIGatewayClient) ListRestAPIs(ctx context.Context) ([]RestAPIInfo, error) {
	var apis []RestAPIInfo
	paginator := apigateway.NewGetRestApisPaginator(c.client, &apigateway.GetRestApisInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list REST APIs: %w", err)
		}

		for _, api := range output.Items {
			endpointType := "EDGE"
			if api.EndpointConfiguration != nil {
				endpointType = string(api.EndpointConfiguration.Types[0])
			}

			apis = append(apis, RestAPIInfo{
				ID:           aws.ToString(api.Id),
				Name:         aws.ToString(api.Name),
				Description:  aws.ToString(api.Description),
				CreatedDate:  aws.ToTime(api.CreatedDate),
				Version:      aws.ToString(api.Version),
				EndpointType: endpointType,
			})
		}
	}

	return apis, nil
}

func (c *APIGatewayClient) ListHTTPAPIs(ctx context.Context) ([]HTTPAPIInfo, error) {
	var apis []HTTPAPIInfo
	var nextToken *string

	for {
		input := &apigatewayv2.GetApisInput{}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := c.clientV2.GetApis(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("unable to list HTTP APIs: %w", err)
		}

		for _, api := range output.Items {
			protocolType := "HTTP"
			if api.ProtocolType != "" {
				protocolType = string(api.ProtocolType)
			}

			apis = append(apis, HTTPAPIInfo{
				APIID:        aws.ToString(api.ApiId),
				Name:         aws.ToString(api.Name),
				ProtocolType: protocolType,
				CreatedDate:  aws.ToTime(api.CreatedDate),
				APIEndpoint:  aws.ToString(api.ApiEndpoint),
				Description:  aws.ToString(api.Description),
			})
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return apis, nil
}
