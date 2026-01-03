package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
)

type ACMClient struct {
	client *acm.Client
}

func NewACMClient(ctx context.Context, profile string) (*ACMClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &ACMClient{
		client: acm.NewFromConfig(cfg),
	}, nil
}

type CertificateInfo struct {
	ARN        string
	DomainName string
	Status     string
	Type       string
	ExpiresAt  *time.Time
}

func (c *ACMClient) ListCertificates(ctx context.Context) ([]CertificateInfo, error) {
	var summaries []CertificateInfo
	paginator := acm.NewListCertificatesPaginator(c.client, &acm.ListCertificatesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to list certificates: %w", err)
		}

		for _, certSummary := range page.CertificateSummaryList {
			summaries = append(summaries, CertificateInfo{
				ARN:        aws.ToString(certSummary.CertificateArn),
				DomainName: aws.ToString(certSummary.DomainName),
			})
		}
	}

	// Fetch details in parallel to avoid "freezing" UI for long periods
	var wg sync.WaitGroup
	resultChan := make(chan CertificateInfo, len(summaries))

	// Limit concurrency to avoid hitting AWS rate limits too hard
	sem := make(chan struct{}, 10)

	for _, s := range summaries {
		wg.Add(1)
		go func(summary CertificateInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			desc, err := c.client.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
				CertificateArn: aws.String(summary.ARN),
			})
			if err != nil {
				summary.Status = "UNKNOWN"
				resultChan <- summary
				return
			}

			cert := desc.Certificate
			resultChan <- CertificateInfo{
				ARN:        aws.ToString(cert.CertificateArn),
				DomainName: aws.ToString(cert.DomainName),
				Status:     string(cert.Status),
				Type:       string(cert.Type),
				ExpiresAt:  cert.NotAfter,
			}
		}(s)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var certificates []CertificateInfo
	for cert := range resultChan {
		certificates = append(certificates, cert)
	}

	return certificates, nil
}
