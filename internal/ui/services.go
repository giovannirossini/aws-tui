package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// featureIcons is the single source of truth for all service names and their icons
var featureIcons = map[string]string{
	"Simple Storage Service (S3)":        "󱐖 ",
	"IAM Users":                          " ",
	"Virtual Private Cloud (VPC)":        "󰛳 ",
	"Lambda Functions":                   "󰘧 ",
	"Elastic Compute Cloud (EC2)":        " ",
	"Relational Database Service (RDS)":  "󰆼 ",
	"CloudWatch":                         "󱖉 ",
	"CloudFront":                         "󰇄 ",
	"ElastiCache (Redis)":                "󰓡 ",
	"Managed Streaming for Kakfa (MSK)":  "󰒔 ",
	"Simple Queue Service (SQS)":         "󰒔 ",
	"Secrets Manager":                    "󰌆 ",
	"Route 53":                           "󰇧 ",
	"Certificate Manager (ACM)":          "󰔕 ",
	"Simple Notification Service (SNS)":  "󰰓 ",
	"KMS Keys":                           "󰌆 ",
	"Data Migration Service (DMS)":       "󰆼 ",
	"Elastic Container Service (ECS)":    "󰙨 ",
	"Billing & Costs":                    "󰠶 ",
	"Security Hub":                       "󰒙 ",
	"Web Application Firewall (WAFv2)":   "󰖛 ",
	"Elastic Container Repository (ECR)": "󰙨 ",
	"Elastic File System (EFS)":          "󰙨 ",
	"AWS Backup":                         "󰁯 ",
	"DynamoDB":                           "󰆼 ",
	"AWS Transfer":                       "󰛳 ",
	"API Gateway":                        "󰓡 ",
}

// serviceHandler is a function type that handles service selection
type serviceHandler func(m *Model) (tea.Model, tea.Cmd)

// getServiceHandlers returns a map of service names to their handlers
func getServiceHandlers() map[string]serviceHandler {
	return map[string]serviceHandler{
		"Simple Storage Service (S3)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewS3
			m.s3Model = NewS3Model(m.selectedProfile, m.styles, m.cache)
			m.s3Model.SetSize(m.width, m.height)
			return *m, m.s3Model.Init()
		},
		"IAM Users": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewIAM
			m.iamModel = NewIAMModel(m.selectedProfile, m.styles, m.cache)
			m.iamModel.SetSize(m.width, m.height)
			return *m, m.iamModel.Init()
		},
		"Virtual Private Cloud (VPC)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewVPC
			m.vpcModel = NewVPCModel(m.selectedProfile, m.styles, m.cache)
			m.vpcModel.SetSize(m.width, m.height)
			return *m, m.vpcModel.Init()
		},
		"Lambda Functions": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewLambda
			m.lambdaModel = NewLambdaModel(m.selectedProfile, m.styles, m.cache)
			m.lambdaModel.SetSize(m.width, m.height)
			return *m, m.lambdaModel.Init()
		},
		"Elastic Compute Cloud (EC2)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewEC2
			m.ec2Model = NewEC2Model(m.selectedProfile, m.styles, m.cache)
			m.ec2Model.SetSize(m.width, m.height)
			return *m, m.ec2Model.Init()
		},
		"Relational Database Service (RDS)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewRDS
			m.rdsModel = NewRDSModel(m.selectedProfile, m.styles, m.cache)
			m.rdsModel.SetSize(m.width, m.height)
			return *m, m.rdsModel.Init()
		},
		"CloudWatch": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewCW
			m.cwModel = NewCWModel(m.selectedProfile, m.styles, m.cache)
			m.cwModel.SetSize(m.width, m.height)
			return *m, m.cwModel.Init()
		},
		"CloudFront": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewCF
			m.cfModel = NewCFModel(m.selectedProfile, m.styles, m.cache)
			m.cfModel.SetSize(m.width, m.height)
			return *m, m.cfModel.Init()
		},
		"ElastiCache (Redis)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewElastiCache
			m.elasticacheModel = NewElastiCacheModel(m.selectedProfile, m.styles, m.cache)
			m.elasticacheModel.SetSize(m.width, m.height)
			return *m, m.elasticacheModel.Init()
		},
		"Managed Streaming for Kakfa (MSK)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewMSK
			m.mskModel = NewMSKModel(m.selectedProfile, m.styles, m.cache)
			m.mskModel.SetSize(m.width, m.height)
			return *m, m.mskModel.Init()
		},
		"Simple Queue Service (SQS)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewSQS
			m.sqsModel = NewSQSModel(m.selectedProfile, m.styles, m.cache)
			m.sqsModel.SetSize(m.width, m.height)
			return *m, m.sqsModel.Init()
		},
		"Secrets Manager": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewSM
			m.smModel = NewSMModel(m.selectedProfile, m.styles, m.cache)
			m.smModel.SetSize(m.width, m.height)
			return *m, m.smModel.Init()
		},
		"Route 53": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewRoute53
			m.route53Model = NewRoute53Model(m.selectedProfile, m.styles, m.cache)
			m.route53Model.SetSize(m.width, m.height)
			return *m, m.route53Model.Init()
		},
		"Certificate Manager (ACM)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewACM
			m.acmModel = NewACMModel(m.selectedProfile, m.styles, m.cache)
			m.acmModel.SetSize(m.width, m.height)
			return *m, m.acmModel.Init()
		},
		"Simple Notification Service (SNS)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewSNS
			m.snsModel = NewSNSModel(m.selectedProfile, m.styles, m.cache)
			m.snsModel.SetSize(m.width, m.height)
			return *m, m.snsModel.Init()
		},
		"KMS Keys": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewKMS
			m.kmsModel = NewKMSModel(m.selectedProfile, m.styles, m.cache)
			m.kmsModel.SetSize(m.width, m.height)
			return *m, m.kmsModel.Init()
		},
		"Data Migration Service (DMS)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewDMS
			m.dmsModel = NewDMSModel(m.selectedProfile, m.styles, m.cache)
			m.dmsModel.SetSize(m.width, m.height)
			return *m, m.dmsModel.Init()
		},
		"Elastic Container Service (ECS)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewECS
			m.ecsModel = NewECSModel(m.selectedProfile, m.styles, m.cache)
			m.ecsModel.SetSize(m.width, m.height)
			return *m, m.ecsModel.Init()
		},
		"Billing & Costs": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewBilling
			m.billingModel = NewBillingModel(m.selectedProfile, m.styles, m.cache)
			m.billingModel.SetSize(m.width, m.height)
			return *m, m.billingModel.Init()
		},
		"Security Hub": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewSecurityHub
			m.securityhubModel = NewSecurityHubModel(m.selectedProfile, m.styles, m.cache)
			m.securityhubModel.SetSize(m.width, m.height)
			return *m, m.securityhubModel.Init()
		},
		"Web Application Firewall (WAFv2)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewWAF
			region := "us-east-1"
			if m.identity != nil {
				region = m.identity.Region
			}
			m.wafModel = NewWAFModel(m.selectedProfile, m.styles, m.cache, region)
			m.wafModel.SetSize(m.width, m.height)
			return *m, m.wafModel.Init()
		},
		"Elastic Container Repository (ECR)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewECR
			m.ecrModel = NewECRModel(m.selectedProfile, m.styles, m.cache)
			m.ecrModel.SetSize(m.width, m.height)
			return *m, m.ecrModel.Init()
		},
		"Elastic File System (EFS)": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewEFS
			m.efsModel = NewEFSModel(m.selectedProfile, m.styles, m.cache)
			m.efsModel.SetSize(m.width, m.height)
			return *m, m.efsModel.Init()
		},
		"AWS Backup": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewBackup
			m.backupModel = NewBackupModel(m.selectedProfile, m.styles, m.cache)
			m.backupModel.SetSize(m.width, m.height)
			return *m, m.backupModel.Init()
		},
		"DynamoDB": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewDynamoDB
			m.dynamodbModel = NewDynamoDBModel(m.selectedProfile, m.styles, m.cache)
			m.dynamodbModel.SetSize(m.width, m.height)
			return *m, m.dynamodbModel.Init()
		},
		"AWS Transfer": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewTransfer
			m.transferModel = NewTransferModel(m.selectedProfile, m.styles, m.cache)
			m.transferModel.SetSize(m.width, m.height)
			return *m, m.transferModel.Init()
		},
		"API Gateway": func(m *Model) (tea.Model, tea.Cmd) {
			m.view = viewAPIGateway
			m.apiGatewayModel = NewAPIGatewayModel(m.selectedProfile, m.styles, m.cache)
			m.apiGatewayModel.SetSize(m.width, m.height)
			return *m, m.apiGatewayModel.Init()
		},
	}
}

// getServiceCategories returns services organized by category
func getServiceCategories() []ServiceCategory {
	return []ServiceCategory{
		{
			Name: "Compute & Containers",
			Services: []string{
				"Elastic Compute Cloud (EC2)",
				"Lambda Functions",
				"Elastic Container Service (ECS)",
				"Elastic Container Repository (ECR)",
			},
		},
		{
			Name: "Storage",
			Services: []string{
				"Simple Storage Service (S3)",
				"Elastic File System (EFS)",
				"AWS Backup",
				"AWS Transfer",
			},
		},
		{
			Name: "Database",
			Services: []string{
				"Relational Database Service (RDS)",
				"DynamoDB",
				"ElastiCache (Redis)",
			},
		},
		{
			Name: "Networking & Content Delivery",
			Services: []string{
				"Virtual Private Cloud (VPC)",
				"Route 53",
				"CloudFront",
				"API Gateway",
			},
		},
		{
			Name: "Security, Identity & Compliance",
			Services: []string{
				"IAM Users",
				"Secrets Manager",
				"Certificate Manager (ACM)",
				"KMS Keys",
				"Web Application Firewall (WAFv2)",
				"Security Hub",
			},
		},
		{
			Name: "Messaging & Integration",
			Services: []string{
				"Simple Notification Service (SNS)",
				"Simple Queue Service (SQS)",
				"Managed Streaming for Kakfa (MSK)",
			},
		},
		{
			Name: "Management & Governance",
			Services: []string{
				"CloudWatch",
				"Billing & Costs",
				"Data Migration Service (DMS)",
			},
		},
	}
}
