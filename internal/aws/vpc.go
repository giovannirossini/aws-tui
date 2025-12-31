package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type EC2Client struct {
	client *ec2.Client
}

func NewEC2Client(ctx context.Context, profile string) (*EC2Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &EC2Client{
		client: ec2.NewFromConfig(cfg),
	}, nil
}

type VPCInfo struct {
	ID        string
	CidrBlock string
	State     string
	IsDefault bool
	Name      string
}

func (c *EC2Client) ListVpcs(ctx context.Context) ([]VPCInfo, error) {
	output, err := c.client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list VPCs: %w", err)
	}

	vpcs := make([]VPCInfo, len(output.Vpcs))
	for i, v := range output.Vpcs {
		name := ""
		for _, tag := range v.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}
		vpcs[i] = VPCInfo{
			ID:        aws.ToString(v.VpcId),
			CidrBlock: aws.ToString(v.CidrBlock),
			State:     string(v.State),
			IsDefault: aws.ToBool(v.IsDefault),
			Name:      name,
		}
	}

	return vpcs, nil
}

type SubnetInfo struct {
	ID               string
	VpcID            string
	CidrBlock        string
	AvailabilityZone string
	State            string
	Name             string
}

func (c *EC2Client) ListSubnets(ctx context.Context) ([]SubnetInfo, error) {
	output, err := c.client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list subnets: %w", err)
	}

	subnets := make([]SubnetInfo, len(output.Subnets))
	for i, s := range output.Subnets {
		name := ""
		for _, tag := range s.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}
		subnets[i] = SubnetInfo{
			ID:               aws.ToString(s.SubnetId),
			VpcID:            aws.ToString(s.VpcId),
			CidrBlock:        aws.ToString(s.CidrBlock),
			AvailabilityZone: aws.ToString(s.AvailabilityZone),
			State:            string(s.State),
			Name:             name,
		}
	}

	return subnets, nil
}

type NatGatewayInfo struct {
	ID        string
	VpcID     string
	State     string
	PublicIP  string
	PrivateIP string
	Name      string
}

func (c *EC2Client) ListNatGateways(ctx context.Context) ([]NatGatewayInfo, error) {
	output, err := c.client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list NAT gateways: %w", err)
	}

	nats := make([]NatGatewayInfo, len(output.NatGateways))
	for i, n := range output.NatGateways {
		name := ""
		for _, tag := range n.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}
		publicIP := ""
		privateIP := ""
		if len(n.NatGatewayAddresses) > 0 {
			publicIP = aws.ToString(n.NatGatewayAddresses[0].PublicIp)
			privateIP = aws.ToString(n.NatGatewayAddresses[0].PrivateIp)
		}
		nats[i] = NatGatewayInfo{
			ID:        aws.ToString(n.NatGatewayId),
			VpcID:     aws.ToString(n.VpcId),
			State:     string(n.State),
			PublicIP:  publicIP,
			PrivateIP: privateIP,
			Name:      name,
		}
	}

	return nats, nil
}

type RouteTableInfo struct {
	ID    string
	VpcID string
	Name  string
}

func (c *EC2Client) ListRouteTables(ctx context.Context) ([]RouteTableInfo, error) {
	output, err := c.client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list route tables: %w", err)
	}

	rts := make([]RouteTableInfo, len(output.RouteTables))
	for i, r := range output.RouteTables {
		name := ""
		for _, tag := range r.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}
		rts[i] = RouteTableInfo{
			ID:    aws.ToString(r.RouteTableId),
			VpcID: aws.ToString(r.VpcId),
			Name:  name,
		}
	}

	return rts, nil
}

type VpnGatewayInfo struct {
	ID    string
	State string
	Type  string
	Name  string
}

func (c *EC2Client) ListVpnGateways(ctx context.Context) ([]VpnGatewayInfo, error) {
	output, err := c.client.DescribeVpnGateways(ctx, &ec2.DescribeVpnGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list VPN gateways: %w", err)
	}

	vpns := make([]VpnGatewayInfo, len(output.VpnGateways))
	for i, v := range output.VpnGateways {
		name := ""
		for _, tag := range v.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = aws.ToString(tag.Value)
				break
			}
		}
		vpns[i] = VpnGatewayInfo{
			ID:    aws.ToString(v.VpnGatewayId),
			State: string(v.State),
			Type:  string(v.Type),
			Name:  name,
		}
	}

	return vpns, nil
}
