package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/efs"
)

type EFSClient struct {
	client *efs.Client
}

func NewEFSClient(ctx context.Context, profile string) (*EFSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &EFSClient{
		client: efs.NewFromConfig(cfg),
	}, nil
}

type FileSystemInfo struct {
	FileSystemId         string
	Name                 string
	CreationTime         time.Time
	LifeCycleState       string
	NumberOfMountTargets int32
	SizeInBytes          int64
	PerformanceMode      string
	ThroughputMode       string
}

func (c *EFSClient) ListFileSystems(ctx context.Context) ([]FileSystemInfo, error) {
	var fileSystems []FileSystemInfo
	paginator := efs.NewDescribeFileSystemsPaginator(c.client, &efs.DescribeFileSystemsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list file systems: %w", err)
		}

		for _, fs := range output.FileSystems {
			name := ""
			for _, tag := range fs.Tags {
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
					break
				}
			}

			info := FileSystemInfo{
				FileSystemId:         aws.ToString(fs.FileSystemId),
				Name:                 name,
				CreationTime:         aws.ToTime(fs.CreationTime),
				LifeCycleState:       string(fs.LifeCycleState),
				NumberOfMountTargets: fs.NumberOfMountTargets,
				PerformanceMode:      string(fs.PerformanceMode),
				ThroughputMode:       string(fs.ThroughputMode),
			}
			if fs.SizeInBytes != nil {
				info.SizeInBytes = fs.SizeInBytes.Value
			}
			fileSystems = append(fileSystems, info)
		}
	}

	return fileSystems, nil
}

type MountTargetInfo struct {
	MountTargetId        string
	FileSystemId         string
	SubnetId             string
	LifeCycleState       string
	IpAddress            string
	NetworkInterfaceId   string
	AvailabilityZoneId   string
	AvailabilityZoneName string
}

func (c *EFSClient) ListMountTargets(ctx context.Context, fileSystemId string) ([]MountTargetInfo, error) {
	output, err := c.client.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
		FileSystemId: aws.String(fileSystemId),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list mount targets: %w", err)
	}

	var targets []MountTargetInfo
	for _, mt := range output.MountTargets {
		targets = append(targets, MountTargetInfo{
			MountTargetId:        aws.ToString(mt.MountTargetId),
			FileSystemId:         aws.ToString(mt.FileSystemId),
			SubnetId:             aws.ToString(mt.SubnetId),
			LifeCycleState:       string(mt.LifeCycleState),
			IpAddress:            aws.ToString(mt.IpAddress),
			NetworkInterfaceId:   aws.ToString(mt.NetworkInterfaceId),
			AvailabilityZoneId:   aws.ToString(mt.AvailabilityZoneId),
			AvailabilityZoneName: aws.ToString(mt.AvailabilityZoneName),
		})
	}

	return targets, nil
}
