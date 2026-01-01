package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManagerClient struct {
	client *secretsmanager.Client
}

func NewSecretsManagerClient(ctx context.Context, profile string) (*SecretsManagerClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &SecretsManagerClient{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}

type SecretInfo struct {
	Name        string
	ARN         string
	Description string
	LastChanged *time.Time
	LastRotated *time.Time
}

func (c *SecretsManagerClient) GetSecretValue(ctx context.Context, secretID string) (string, error) {
	output, err := c.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return "", fmt.Errorf("unable to get secret value: %w", err)
	}

	if output.SecretString != nil {
		return *output.SecretString, nil
	}

	return "Binary secret values are not supported yet", nil
}

func (c *SecretsManagerClient) ListSecrets(ctx context.Context) ([]SecretInfo, error) {
	output, err := c.client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list secrets: %w", err)
	}

	var secrets []SecretInfo
	for _, s := range output.SecretList {
		secrets = append(secrets, SecretInfo{
			Name:        aws.ToString(s.Name),
			ARN:         aws.ToString(s.ARN),
			Description: aws.ToString(s.Description),
			LastChanged: s.LastChangedDate,
			LastRotated: s.LastRotatedDate,
		})
	}

	return secrets, nil
}
