package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/backup"
)

type BackupClient struct {
	client *backup.Client
}

func NewBackupClient(ctx context.Context, profile string) (*BackupClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &BackupClient{
		client: backup.NewFromConfig(cfg),
	}, nil
}

type BackupPlanInfo struct {
	BackupPlanId      string
	BackupPlanArn     string
	BackupPlanName    string
	CreationDate      time.Time
	LastExecutionDate time.Time
	VersionId         string
}

func (c *BackupClient) ListBackupPlans(ctx context.Context) ([]BackupPlanInfo, error) {
	var plans []BackupPlanInfo
	paginator := backup.NewListBackupPlansPaginator(c.client, &backup.ListBackupPlansInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list backup plans: %w", err)
		}

		for _, p := range output.BackupPlansList {
			plans = append(plans, BackupPlanInfo{
				BackupPlanId:      aws.ToString(p.BackupPlanId),
				BackupPlanArn:     aws.ToString(p.BackupPlanArn),
				BackupPlanName:    aws.ToString(p.BackupPlanName),
				CreationDate:      aws.ToTime(p.CreationDate),
				LastExecutionDate: aws.ToTime(p.LastExecutionDate),
				VersionId:         aws.ToString(p.VersionId),
			})
		}
	}

	return plans, nil
}

type BackupJobInfo struct {
	BackupJobId       string
	BackupVaultName   string
	BackupVaultArn    string
	ResourceArn       string
	ResourceType      string
	State             string
	PercentDone       string
	BackupSizeInBytes int64
	CreationDate      time.Time
	CompletionDate    time.Time
}

func (c *BackupClient) ListBackupJobs(ctx context.Context) ([]BackupJobInfo, error) {
	var jobs []BackupJobInfo
	paginator := backup.NewListBackupJobsPaginator(c.client, &backup.ListBackupJobsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list backup jobs: %w", err)
		}

		for _, j := range output.BackupJobs {
			jobs = append(jobs, BackupJobInfo{
				BackupJobId:       aws.ToString(j.BackupJobId),
				BackupVaultName:   aws.ToString(j.BackupVaultName),
				BackupVaultArn:    aws.ToString(j.BackupVaultArn),
				ResourceArn:       aws.ToString(j.ResourceArn),
				ResourceType:      aws.ToString(j.ResourceType),
				State:             string(j.State),
				PercentDone:       aws.ToString(j.PercentDone),
				BackupSizeInBytes: aws.ToInt64(j.BackupSizeInBytes),
				CreationDate:      aws.ToTime(j.CreationDate),
				CompletionDate:    aws.ToTime(j.CompletionDate),
			})
		}
	}

	return jobs, nil
}
