package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type EC2ResourcesClient struct {
	ec2Client *ec2.Client
	elbClient *elasticloadbalancingv2.Client
}

func NewEC2ResourcesClient(ctx context.Context, profile string) (*EC2ResourcesClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &EC2ResourcesClient{
		ec2Client: ec2.NewFromConfig(cfg),
		elbClient: elasticloadbalancingv2.NewFromConfig(cfg),
	}, nil
}

type InstanceInfo struct {
	ID               string
	Type             string
	State            string
	PublicIP         string
	PrivateIP        string
	AvailabilityZone string
	Name             string
}

func (c *EC2ResourcesClient) ListInstances(ctx context.Context) ([]InstanceInfo, error) {
	output, err := c.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list instances: %w", err)
	}

	var instances []InstanceInfo
	for _, reservation := range output.Reservations {
		for _, i := range reservation.Instances {
			name := ""
			for _, tag := range i.Tags {
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
					break
				}
			}
			instances = append(instances, InstanceInfo{
				ID:               aws.ToString(i.InstanceId),
				Type:             string(i.InstanceType),
				State:            string(i.State.Name),
				PublicIP:         aws.ToString(i.PublicIpAddress),
				PrivateIP:        aws.ToString(i.PrivateIpAddress),
				AvailabilityZone: aws.ToString(i.Placement.AvailabilityZone),
				Name:             name,
			})
		}
	}

	return instances, nil
}

type SecurityGroupInfo struct {
	ID          string
	Name        string
	Description string
	VpcID       string
}

func (c *EC2ResourcesClient) ListSecurityGroups(ctx context.Context) ([]SecurityGroupInfo, error) {
	output, err := c.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list security groups: %w", err)
	}

	sgs := make([]SecurityGroupInfo, len(output.SecurityGroups))
	for i, s := range output.SecurityGroups {
		sgs[i] = SecurityGroupInfo{
			ID:          aws.ToString(s.GroupId),
			Name:        aws.ToString(s.GroupName),
			Description: aws.ToString(s.Description),
			VpcID:       aws.ToString(s.VpcId),
		}
	}

	return sgs, nil
}

type VolumeInfo struct {
	ID               string
	Size             int32
	Type             string
	State            string
	AvailabilityZone string
	InstanceID       string
	Name             string
}

func (c *EC2ResourcesClient) ListVolumes(ctx context.Context) ([]VolumeInfo, error) {
	output, err := c.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list volumes: %w", err)
	}

	volumes := make([]VolumeInfo, len(output.Volumes))
	for i, v := range output.Volumes {
		name := ""
		for _, tag := range v.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}
		instanceID := ""
		if len(v.Attachments) > 0 {
			instanceID = aws.ToString(v.Attachments[0].InstanceId)
		}
		volumes[i] = VolumeInfo{
			ID:               aws.ToString(v.VolumeId),
			Size:             aws.ToInt32(v.Size),
			Type:             string(v.VolumeType),
			State:            string(v.State),
			AvailabilityZone: aws.ToString(v.AvailabilityZone),
			InstanceID:       instanceID,
			Name:             name,
		}
	}

	return volumes, nil
}

type TargetGroupInfo struct {
	ARN        string
	Name       string
	Protocol   string
	Port       int32
	VpcID      string
	TargetType string
}

func (c *EC2ResourcesClient) ListTargetGroups(ctx context.Context) ([]TargetGroupInfo, error) {
	output, err := c.elbClient.DescribeTargetGroups(ctx, &elasticloadbalancingv2.DescribeTargetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list target groups: %w", err)
	}

	tgs := make([]TargetGroupInfo, len(output.TargetGroups))
	for i, t := range output.TargetGroups {
		tgs[i] = TargetGroupInfo{
			ARN:        aws.ToString(t.TargetGroupArn),
			Name:       aws.ToString(t.TargetGroupName),
			Protocol:   string(t.Protocol),
			Port:       aws.ToInt32(t.Port),
			VpcID:      aws.ToString(t.VpcId),
			TargetType: string(t.TargetType),
		}
	}

	return tgs, nil
}
