package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

type KMSClient struct {
	client *kms.Client
}

func NewKMSClient(ctx context.Context, profile string) (*KMSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &KMSClient{
		client: kms.NewFromConfig(cfg),
	}, nil
}

type KMSKeyInfo struct {
	ID          string
	ARN         string
	Alias       string
	Description string
	State       string
	Enabled     bool
	Manager     string // AWS or CUSTOMER
}

func (c *KMSClient) ListKeys(ctx context.Context) ([]KMSKeyInfo, error) {
	// 1. Fetch all aliases first to map them to keys
	aliases := make(map[string][]string) // KeyID -> AliasNames
	aliasPaginator := kms.NewListAliasesPaginator(c.client, &kms.ListAliasesInput{})
	for aliasPaginator.HasMorePages() {
		page, err := aliasPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list aliases: %w", err)
		}
		for _, a := range page.Aliases {
			if a.TargetKeyId != nil {
				aliases[*a.TargetKeyId] = append(aliases[*a.TargetKeyId], aws.ToString(a.AliasName))
			}
		}
	}

	// 2. List all keys
	var keyIds []string
	keyPaginator := kms.NewListKeysPaginator(c.client, &kms.ListKeysInput{})
	for keyPaginator.HasMorePages() {
		page, err := keyPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list keys: %w", err)
		}
		for _, k := range page.Keys {
			keyIds = append(keyIds, aws.ToString(k.KeyId))
		}
	}

	// 3. Fetch details in parallel
	var wg sync.WaitGroup
	resultChan := make(chan KMSKeyInfo, len(keyIds))
	sem := make(chan struct{}, 10)

	for _, id := range keyIds {
		wg.Add(1)
		go func(keyId string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			desc, err := c.client.DescribeKey(ctx, &kms.DescribeKeyInput{
				KeyId: aws.String(keyId),
			})

			if err != nil {
				resultChan <- KMSKeyInfo{ID: keyId, State: "UNKNOWN"}
				return
			}

			k := desc.KeyMetadata
			aliasList := aliases[keyId]
			primaryAlias := ""
			if len(aliasList) > 0 {
				primaryAlias = aliasList[0]
				// Prefer the one that doesn't start with "alias/aws/" if multiple exist
				for _, a := range aliasList {
					if !strings.HasPrefix(a, "alias/aws/") {
						primaryAlias = a
						break
					}
				}
			}

			manager := "CUSTOMER"
			if k.KeyManager == "AWS" {
				manager = "AWS"
			}

			resultChan <- KMSKeyInfo{
				ID:          aws.ToString(k.KeyId),
				ARN:         aws.ToString(k.Arn),
				Alias:       primaryAlias,
				Description: aws.ToString(k.Description),
				State:       string(k.KeyState),
				Enabled:     k.Enabled,
				Manager:     manager,
			}
		}(id)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var keys []KMSKeyInfo
	for k := range resultChan {
		keys = append(keys, k)
	}

	return keys, nil
}
