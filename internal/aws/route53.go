package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

type Route53Client struct {
	client *route53.Client
}

func NewRoute53Client(ctx context.Context, profile string) (*Route53Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &Route53Client{
		client: route53.NewFromConfig(cfg),
	}, nil
}

type HostedZoneInfo struct {
	ID          string
	Name        string
	RecordCount int64
	IsPrivate   bool
	Comment     string
}

func (c *Route53Client) ListHostedZones(ctx context.Context) ([]HostedZoneInfo, error) {
	output, err := c.client.ListHostedZones(ctx, &route53.ListHostedZonesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list hosted zones: %w", err)
	}

	var zones []HostedZoneInfo
	for _, z := range output.HostedZones {
		comment := ""
		if z.Config != nil && z.Config.Comment != nil {
			comment = *z.Config.Comment
		}

		zones = append(zones, HostedZoneInfo{
			ID:          strings.TrimPrefix(*z.Id, "/hostedzone/"),
			Name:        aws.ToString(z.Name),
			RecordCount: aws.ToInt64(z.ResourceRecordSetCount),
			IsPrivate:   z.Config != nil && z.Config.PrivateZone,
			Comment:     comment,
		})
	}

	return zones, nil
}

type ResourceRecordSetInfo struct {
	Name   string
	Type   string
	TTL    int64
	Values []string
	Alias  string
}

func (c *Route53Client) ListResourceRecordSets(ctx context.Context, zoneID string) ([]ResourceRecordSetInfo, error) {
	output, err := c.client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list record sets: %w", err)
	}

	var records []ResourceRecordSetInfo
	for _, r := range output.ResourceRecordSets {
		var values []string
		for _, v := range r.ResourceRecords {
			values = append(values, aws.ToString(v.Value))
		}

		alias := ""
		if r.AliasTarget != nil {
			alias = aws.ToString(r.AliasTarget.DNSName)
		}

		records = append(records, ResourceRecordSetInfo{
			Name:   aws.ToString(r.Name),
			Type:   string(r.Type),
			TTL:    aws.ToInt64(r.TTL),
			Values: values,
			Alias:  alias,
		})
	}

	return records, nil
}
