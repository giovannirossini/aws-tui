package aws

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type ECRClient struct {
	client *ecr.Client
}

func NewECRClient(ctx context.Context, profile string) (*ECRClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &ECRClient{
		client: ecr.NewFromConfig(cfg),
	}, nil
}

type RepositoryInfo struct {
	Name            string
	RegistryId      string
	RepositoryUri   string
	CreatedAt       time.Time
	ImageTagMut     string
	ScanOnPush      bool
	EncryptionType  string
}

func (c *ECRClient) ListRepositories(ctx context.Context) ([]RepositoryInfo, error) {
	var repos []RepositoryInfo
	paginator := ecr.NewDescribeRepositoriesPaginator(c.client, &ecr.DescribeRepositoriesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list repositories: %w", err)
		}

		for _, r := range output.Repositories {
			repo := RepositoryInfo{
				Name:          aws.ToString(r.RepositoryName),
				RegistryId:    aws.ToString(r.RegistryId),
				RepositoryUri: aws.ToString(r.RepositoryUri),
				CreatedAt:     aws.ToTime(r.CreatedAt),
				ImageTagMut:   string(r.ImageTagMutability),
			}
			if r.ImageScanningConfiguration != nil {
				repo.ScanOnPush = r.ImageScanningConfiguration.ScanOnPush
			}
			if r.EncryptionConfiguration != nil {
				repo.EncryptionType = string(r.EncryptionConfiguration.EncryptionType)
			}
			repos = append(repos, repo)
		}
	}

	return repos, nil
}

type ImageInfo struct {
	Tags           []string
	Digest         string
	PushedAt       time.Time
	Size           int64
	Status         string
	ArtifactMediaType string
}

func (c *ECRClient) ListImages(ctx context.Context, repositoryName string) ([]ImageInfo, error) {
	var images []ImageInfo
	paginator := ecr.NewDescribeImagesPaginator(c.client, &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repositoryName),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list images: %w", err)
		}

		for _, i := range output.ImageDetails {
			image := ImageInfo{
				Tags:              i.ImageTags,
				Digest:            aws.ToString(i.ImageDigest),
				PushedAt:          aws.ToTime(i.ImagePushedAt),
				Size:              aws.ToInt64(i.ImageSizeInBytes),
				ArtifactMediaType: aws.ToString(i.ArtifactMediaType),
			}
			if i.ImageScanStatus != nil {
				image.Status = string(i.ImageScanStatus.Status)
			}
			images = append(images, image)
		}
	}

	// Sort by PushedAt descending
	sort.Slice(images, func(i, j int) bool {
		return images[i].PushedAt.After(images[j].PushedAt)
	})

	return images, nil
}
