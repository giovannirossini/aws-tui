package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IAMClient struct {
	client *iam.Client
}

func NewIAMClient(ctx context.Context, profile string) (*IAMClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	// IAM is global but the SDK often needs a region to be set to use the default endpoint resolver
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	return &IAMClient{
		client: iam.NewFromConfig(cfg),
	}, nil
}

type IAMUserInfo struct {
	UserName            string
	UserID              string
	Path                string
	Arn                 string
	CreateDate          time.Time
	PasswordLastUsed    *time.Time
	PasswordExists      bool
	MFAEnabled          bool
	AccessKeysCount     int
	PasswordLastChanged *time.Time
}

type AccessKeyInfo struct {
	AccessKeyId string
	Status      string
	CreateDate  time.Time
}

func (c *IAMClient) ListUsers(ctx context.Context) ([]IAMUserInfo, error) {
	output, err := c.client.ListUsers(ctx, &iam.ListUsersInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list users: %w", err)
	}

	users := make([]IAMUserInfo, len(output.Users))
	for i, u := range output.Users {
		users[i] = IAMUserInfo{
			UserName:         aws.ToString(u.UserName),
			UserID:           aws.ToString(u.UserId),
			Path:             aws.ToString(u.Path),
			Arn:              aws.ToString(u.Arn),
			CreateDate:       aws.ToTime(u.CreateDate),
			PasswordLastUsed: u.PasswordLastUsed,
		}
	}

	return users, nil
}

func (c *IAMClient) GetUserDetails(ctx context.Context, userName string) (*IAMUserInfo, []AccessKeyInfo, error) {
	// Get User Basic Info
	uOut, err := c.client.GetUser(ctx, &iam.GetUserInput{UserName: aws.String(userName)})
	if err != nil {
		return nil, nil, err
	}
	u := uOut.User

	info := &IAMUserInfo{
		UserName:         aws.ToString(u.UserName),
		UserID:           aws.ToString(u.UserId),
		CreateDate:       aws.ToTime(u.CreateDate),
		PasswordLastUsed: u.PasswordLastUsed,
	}

	// Check if login profile exists (Console access)
	_, err = c.client.GetLoginProfile(ctx, &iam.GetLoginProfileInput{UserName: aws.String(userName)})
	if err == nil {
		info.PasswordExists = true
	}

	// Check MFA
	mfaOut, err := c.client.ListMFADevices(ctx, &iam.ListMFADevicesInput{UserName: aws.String(userName)})
	if err == nil && len(mfaOut.MFADevices) > 0 {
		info.MFAEnabled = true
	}

	// List Access Keys
	keysOut, err := c.client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{UserName: aws.String(userName)})
	var keys []AccessKeyInfo
	if err == nil {
		info.AccessKeysCount = len(keysOut.AccessKeyMetadata)
		for _, k := range keysOut.AccessKeyMetadata {
			keys = append(keys, AccessKeyInfo{
				AccessKeyId: aws.ToString(k.AccessKeyId),
				Status:      string(k.Status),
				CreateDate:  aws.ToTime(k.CreateDate),
			})
		}
	}

	return info, keys, nil
}

func (c *IAMClient) CreateUser(ctx context.Context, userName string) error {
	_, err := c.client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	return err
}

func (c *IAMClient) DeleteUser(ctx context.Context, userName string) error {
	_, err := c.client.DeleteUser(ctx, &iam.DeleteUserInput{
		UserName: aws.String(userName),
	})
	return err
}

func (c *IAMClient) CreateLoginProfile(ctx context.Context, userName, password string) error {
	_, err := c.client.CreateLoginProfile(ctx, &iam.CreateLoginProfileInput{
		UserName:              aws.String(userName),
		Password:              aws.String(password),
		PasswordResetRequired: true,
	})
	return err
}

func (c *IAMClient) DeleteLoginProfile(ctx context.Context, userName string) error {
	_, err := c.client.DeleteLoginProfile(ctx, &iam.DeleteLoginProfileInput{
		UserName: aws.String(userName),
	})
	return err
}

func (c *IAMClient) UpdateLoginProfile(ctx context.Context, userName, password string) error {
	_, err := c.client.UpdateLoginProfile(ctx, &iam.UpdateLoginProfileInput{
		UserName:              aws.String(userName),
		Password:              aws.String(password),
		PasswordResetRequired: aws.Bool(true),
	})
	return err
}

func (c *IAMClient) ListAccountAliases(ctx context.Context) ([]string, error) {
	output, err := c.client.ListAccountAliases(ctx, &iam.ListAccountAliasesInput{})
	if err != nil {
		return nil, err
	}
	return output.AccountAliases, nil
}
