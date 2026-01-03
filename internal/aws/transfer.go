package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
)

type TransferClient struct {
	client *transfer.Client
}

func NewTransferClient(ctx context.Context, profile string) (*TransferClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &TransferClient{
		client: transfer.NewFromConfig(cfg),
	}, nil
}

type TransferServerInfo struct {
	ServerId             string
	Arn                  string
	State                string
	EndpointType         string
	IdentityProviderType string
	UserCount            int32
}

func (c *TransferClient) ListServers(ctx context.Context) ([]TransferServerInfo, error) {
	var servers []TransferServerInfo
	paginator := transfer.NewListServersPaginator(c.client, &transfer.ListServersInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list servers: %w", err)
		}

		for _, s := range output.Servers {
			servers = append(servers, TransferServerInfo{
				ServerId:             aws.ToString(s.ServerId),
				Arn:                  aws.ToString(s.Arn),
				State:                string(s.State),
				EndpointType:         string(s.EndpointType),
				IdentityProviderType: string(s.IdentityProviderType),
				UserCount:            aws.ToInt32(s.UserCount),
			})
		}
	}

	return servers, nil
}

type TransferUserInfo struct {
	UserName          string
	Role              string
	HomeDirectory     string
	SshPublicKeyCount int
}

func (c *TransferClient) ListUsers(ctx context.Context, serverId string) ([]TransferUserInfo, error) {
	var users []TransferUserInfo
	paginator := transfer.NewListUsersPaginator(c.client, &transfer.ListUsersInput{
		ServerId: aws.String(serverId),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list users: %w", err)
		}

		for _, u := range output.Users {
			desc, err := c.client.DescribeUser(ctx, &transfer.DescribeUserInput{
				ServerId: aws.String(serverId),
				UserName: u.UserName,
			})
			if err != nil {
				continue
			}

			user := desc.User
			users = append(users, TransferUserInfo{
				UserName:          aws.ToString(user.UserName),
				Role:              aws.ToString(user.Role),
				HomeDirectory:     aws.ToString(user.HomeDirectory),
				SshPublicKeyCount: len(user.SshPublicKeys),
			})
		}
	}

	return users, nil
}
