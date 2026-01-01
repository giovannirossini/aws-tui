package ui

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
	"github.com/sahilm/fuzzy"
)

type focus int

const (
	focusContent focus = iota
)

type viewState int

const (
	viewHome viewState = iota
	viewS3
	viewIAM
	viewVPC
	viewLambda
	viewEC2
	viewRDS
	viewCW
	viewCF
	viewElastiCache
	viewMSK
	viewSQS
	viewSM
	viewRoute53
	viewACM
	viewSNS
	viewKMS
	viewDMS
	viewECS
	viewBilling
	viewSecurityHub
	viewWAF
	viewECR
	viewEFS
	viewBackup
	viewDynamoDB
	viewTransfer
	viewAPIGateway
)

type ServiceCategory struct {
	Name     string
	Services []string
}

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

type Model struct {
	profiles         []string
	selectedProfile  string
	profileSelector  ProfileSelector
	styles           Styles
	focus            focus
	view             viewState
	s3Model          S3Model
	iamModel         IAMModel
	vpcModel         VPCModel
	lambdaModel      LambdaModel
	ec2Model         EC2Model
	rdsModel         RDSModel
	cwModel          CWModel
	cfModel          CFModel
	elasticacheModel ElastiCacheModel
	mskModel         MSKModel
	sqsModel         SQSModel
	smModel          SMModel
	route53Model     Route53Model
	acmModel         ACMModel
	snsModel         SNSModel
	kmsModel         KMSModel
	dmsModel         DMSModel
	ecsModel         ECSModel
	billingModel     BillingModel
	securityhubModel SecurityHubModel
	wafModel         WAFModel
	ecrModel         ECRModel
	efsModel         EFSModel
	backupModel      BackupModel
	dynamodbModel    DynamoDBModel
	transferModel    TransferModel
	apiGatewayModel  APIGatewayModel
	categories       []ServiceCategory
	selectedCategory int
	selectedService  int
	// Search/Filter
	searchInput      textinput.Model
	searching        bool
	filteredServices []string
	selectedFiltered int
	width            int
	height           int
	ready            bool
	identity         *aws.IdentityInfo
	cache            *cache.Cache
	cacheKeys        *cache.KeyBuilder
}

type IdentityMsg *aws.IdentityInfo

func NewModel() (Model, error) {
	profiles, err := aws.GetProfiles()
	if err != nil {
		return Model{}, err
	}

	selected := ""
	// 1. Try to use AWS_PROFILE if set
	if p := os.Getenv("AWS_PROFILE"); p != "" {
		for _, profile := range profiles {
			if profile == p {
				selected = p
				break
			}
		}
	}

	// 2. If no AWS_PROFILE or not found, try "default"
	if selected == "" {
		for _, p := range profiles {
			if p == "default" {
				selected = "default"
				break
			}
		}
	}

	// 3. If "default" not found, use the first in the sorted list
	if selected == "" && len(profiles) > 0 {
		selected = profiles[0]
	}

	// Fallback if no profiles found at all
	if selected == "" {
		selected = "default"
	}

	styles := DefaultStyles()
	ps := NewProfileSelector(profiles, selected, styles)
	appCache := cache.New()

	// Start background cache cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			appCache.CleanExpired()
		}
	}()

	ti := textinput.New()
	ti.Placeholder = "Search services..."
	ti.Prompt = "/ "
	ti.CharLimit = 64
	ti.Width = 30

	return Model{
		profiles:         profiles,
		selectedProfile:  selected,
		profileSelector:  ps,
		styles:           styles,
		focus:            focusContent,
		view:             viewHome,
		categories:       getServiceCategories(),
		selectedCategory: 0,
		selectedService:  0,
		searchInput:      ti,
		cache:            appCache,
		cacheKeys:        cache.NewKeyBuilder(selected),
	}, nil
}

func (m Model) Init() tea.Cmd {
	return m.fetchIdentity()
}

func (m Model) fetchIdentity() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		if cached, ok := m.cache.Get(m.cacheKeys.Identity()); ok {
			if id, ok := cached.(*aws.IdentityInfo); ok {
				return IdentityMsg(id)
			}
		}

		ctx := context.Background()
		stsClient, err := aws.NewSTSClient(ctx, m.selectedProfile)
		if err != nil {
			return nil
		}
		id, err := stsClient.GetCallerIdentity(ctx)
		if err != nil {
			return nil
		}

		// Try to fetch account alias
		iamClient, err := aws.NewIAMClient(ctx, m.selectedProfile)
		if err == nil {
			aliases, err := iamClient.ListAccountAliases(ctx)
			if err == nil && len(aliases) > 0 {
				id.Alias = aliases[0]
			}
		}

		// Cache the identity
		m.cache.Set(m.cacheKeys.Identity(), id, cache.TTLIdentity)

		return IdentityMsg(id)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.profileSelector.SetSize(m.width, m.height)
		if m.view == viewS3 {
			m.s3Model.SetSize(m.width, m.height)
			m.s3Model, cmd = m.s3Model.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewIAM {
			m.iamModel.SetSize(m.width, m.height)
			m.iamModel, cmd = m.iamModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewVPC {
			m.vpcModel.SetSize(m.width, m.height)
			m.vpcModel, cmd = m.vpcModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewLambda {
			m.lambdaModel.SetSize(m.width, m.height)
			m.lambdaModel, cmd = m.lambdaModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewEC2 {
			m.ec2Model.SetSize(m.width, m.height)
			m.ec2Model, cmd = m.ec2Model.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewRDS {
			m.rdsModel.SetSize(m.width, m.height)
			m.rdsModel, cmd = m.rdsModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewCW {
			m.cwModel.SetSize(m.width, m.height)
			m.cwModel, cmd = m.cwModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewCF {
			m.cfModel.SetSize(m.width, m.height)
			m.cfModel, cmd = m.cfModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewElastiCache {
			m.elasticacheModel.SetSize(m.width, m.height)
			m.elasticacheModel, cmd = m.elasticacheModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewMSK {
			m.mskModel.SetSize(m.width, m.height)
			m.mskModel, cmd = m.mskModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewSQS {
			m.sqsModel.SetSize(m.width, m.height)
			m.sqsModel, cmd = m.sqsModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewSM {
			m.smModel.SetSize(m.width, m.height)
			m.smModel, cmd = m.smModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewRoute53 {
			m.route53Model.SetSize(m.width, m.height)
			m.route53Model, cmd = m.route53Model.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewACM {
			m.acmModel.SetSize(m.width, m.height)
			m.acmModel, cmd = m.acmModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewSNS {
			m.snsModel.SetSize(m.width, m.height)
			m.snsModel, cmd = m.snsModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewKMS {
			m.kmsModel.SetSize(m.width, m.height)
			m.kmsModel, cmd = m.kmsModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewDMS {
			m.dmsModel.SetSize(m.width, m.height)
			m.dmsModel, cmd = m.dmsModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewECS {
			m.ecsModel.SetSize(m.width, m.height)
			m.ecsModel, cmd = m.ecsModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewBilling {
			m.billingModel.SetSize(m.width, m.height)
			m.billingModel, cmd = m.billingModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewSecurityHub {
			m.securityhubModel.SetSize(m.width, m.height)
			m.securityhubModel, cmd = m.securityhubModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewWAF {
			m.wafModel.SetSize(m.width, m.height)
			m.wafModel, cmd = m.wafModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewECR {
			m.ecrModel.SetSize(m.width, m.height)
			m.ecrModel, cmd = m.ecrModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewEFS {
			m.efsModel.SetSize(m.width, m.height)
			m.efsModel, cmd = m.efsModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewBackup {
			m.backupModel.SetSize(m.width, m.height)
			m.backupModel, cmd = m.backupModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewDynamoDB {
			m.dynamodbModel.SetSize(m.width, m.height)
			m.dynamodbModel, cmd = m.dynamodbModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewTransfer {
			m.transferModel.SetSize(m.width, m.height)
			m.transferModel, cmd = m.transferModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.view == viewAPIGateway {
			m.apiGatewayModel.SetSize(m.width, m.height)
			m.apiGatewayModel, cmd = m.apiGatewayModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		m.ready = true

	case tea.KeyMsg:
		if m.profileSelector.active {
			m.profileSelector, cmd = m.profileSelector.Update(msg)
			return m, cmd
		}

		// Handle global keys that should work in all views (unless in input state)
		if !m.isInputFocused() {
			switch msg.String() {
			case "/":
				if m.view == viewHome {
					m.searching = true
					m.selectedFiltered = 0
					m.searchInput.Focus()
					m.updateFilter()
					return m, textinput.Blink
				}
			case "p", "P":
				m.profileSelector.active = true
				m.profileSelector.list.FilterInput.Focus()
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		} else {
			// Even in input mode, ctrl+c should quit
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}

			if m.searching {
				switch msg.String() {
				case "esc":
					m.searching = false
					m.searchInput.Blur()
					m.searchInput.SetValue("")
					return m, nil
				case "enter":
					if len(m.filteredServices) > 0 {
						serviceName := m.filteredServices[m.selectedFiltered]
						m.searching = false
						m.searchInput.Blur()
						m.searchInput.SetValue("")
						return m.handleServiceSelection(serviceName)
					}
				case "up":
					if m.selectedFiltered > 0 {
						m.selectedFiltered--
					}
				case "down":
					if m.selectedFiltered < len(m.filteredServices)-1 {
						m.selectedFiltered++
					}
				}

				var tiCmd tea.Cmd
				m.searchInput, tiCmd = m.searchInput.Update(msg)
				m.updateFilter()
				return m, tiCmd
			}
		}

		if m.view == viewS3 {
			// ... (existing S3 handling)
			if msg.String() == "esc" && m.s3Model.state == S3StateBuckets {
				m.view = viewHome
				return m, nil
			}

			// Special handling for edit which requires suspension
			if msg.String() == "e" && m.s3Model.state == S3StateObjects {
				if item, ok := m.s3Model.list.SelectedItem().(s3Item); ok && !item.isFolder && !item.isBucket {
					return m, tea.ExecProcess(m.s3Model.getEditCommand(item.key), func(err error) tea.Msg {
						if err != nil {
							return S3ErrorMsg(err)
						}
						return m.s3Model.uploadEditedFile(item.key)
					})
				}
			}

			m.s3Model, cmd = m.s3Model.Update(msg)
			return m, cmd
		}

		if m.view == viewIAM {
			if msg.String() == "esc" && m.iamModel.state == IAMStateUsers {
				m.view = viewHome
				return m, nil
			}
			m.iamModel, cmd = m.iamModel.Update(msg)
			return m, cmd
		}

		if m.view == viewVPC {
			if msg.String() == "esc" && m.vpcModel.state == VPCStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.vpcModel, cmd = m.vpcModel.Update(msg)
			return m, cmd
		}

		if m.view == viewLambda {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.lambdaModel, cmd = m.lambdaModel.Update(msg)
			return m, cmd
		}

		if m.view == viewEC2 {
			if msg.String() == "esc" && m.ec2Model.state == EC2StateMenu {
				m.view = viewHome
				return m, nil
			}
			m.ec2Model, cmd = m.ec2Model.Update(msg)
			return m, cmd
		}

		if m.view == viewRDS {
			if msg.String() == "esc" && m.rdsModel.state == RDSStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.rdsModel, cmd = m.rdsModel.Update(msg)
			return m, cmd
		}

		if m.view == viewCW {
			if msg.String() == "esc" && m.cwModel.state == CWStateLogStreams && m.cwModel.originView == viewECS {
				m.view = viewECS
				return m, nil
			}
			if msg.String() == "esc" && m.cwModel.state == CWStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.cwModel, cmd = m.cwModel.Update(msg)
			return m, cmd
		}

		if m.view == viewCF {
			if msg.String() == "esc" && m.cfModel.state == CFStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.cfModel, cmd = m.cfModel.Update(msg)
			return m, cmd
		}

		if m.view == viewElastiCache {
			if msg.String() == "esc" && m.elasticacheModel.state == ElastiCacheStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.elasticacheModel, cmd = m.elasticacheModel.Update(msg)
			return m, cmd
		}

		if m.view == viewMSK {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.mskModel, cmd = m.mskModel.Update(msg)
			return m, cmd
		}

		if m.view == viewSQS {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.sqsModel, cmd = m.sqsModel.Update(msg)
			return m, cmd
		}

		if m.view == viewSM {
			if msg.String() == "esc" && m.smModel.state == SMStateSecrets {
				m.view = viewHome
				return m, nil
			}
			m.smModel, cmd = m.smModel.Update(msg)
			return m, cmd
		}

		if m.view == viewRoute53 {
			if msg.String() == "esc" && m.route53Model.state == Route53StateZones {
				m.view = viewHome
				return m, nil
			}
			m.route53Model, cmd = m.route53Model.Update(msg)
			return m, cmd
		}

		if m.view == viewACM {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.acmModel, cmd = m.acmModel.Update(msg)
			return m, cmd
		}

		if m.view == viewSNS {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.snsModel, cmd = m.snsModel.Update(msg)
			return m, cmd
		}

		if m.view == viewKMS {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.kmsModel, cmd = m.kmsModel.Update(msg)
			return m, cmd
		}

		if m.view == viewDMS {
			if msg.String() == "esc" && m.dmsModel.state == DMSStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.dmsModel, cmd = m.dmsModel.Update(msg)
			return m, cmd
		}

		if m.view == viewECS {
			if msg.String() == "esc" && m.ecsModel.state == ECSStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.ecsModel, cmd = m.ecsModel.Update(msg)
			return m, cmd
		}

		if m.view == viewBilling {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.billingModel, cmd = m.billingModel.Update(msg)
			return m, cmd
		}

		if m.view == viewSecurityHub {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.securityhubModel, cmd = m.securityhubModel.Update(msg)
			return m, cmd
		}

		if m.view == viewWAF {
			if msg.String() == "esc" && m.wafModel.state == WAFStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.wafModel, cmd = m.wafModel.Update(msg)
			return m, cmd
		}

		if m.view == viewECR {
			if msg.String() == "esc" && m.ecrModel.state == ECRStateRepositories {
				m.view = viewHome
				return m, nil
			}
			m.ecrModel, cmd = m.ecrModel.Update(msg)
			return m, cmd
		}

		if m.view == viewEFS {
			if msg.String() == "esc" && m.efsModel.state == EFSStateFileSystems {
				m.view = viewHome
				return m, nil
			}
			m.efsModel, cmd = m.efsModel.Update(msg)
			return m, cmd
		}

		if m.view == viewBackup {
			if msg.String() == "esc" && m.backupModel.state == BackupStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.backupModel, cmd = m.backupModel.Update(msg)
			return m, cmd
		}

		if m.view == viewDynamoDB {
			if msg.String() == "esc" {
				m.view = viewHome
				return m, nil
			}
			m.dynamodbModel, cmd = m.dynamodbModel.Update(msg)
			return m, cmd
		}

		if m.view == viewTransfer {
			if msg.String() == "esc" && m.transferModel.state == TransferStateServers {
				m.view = viewHome
				return m, nil
			}
			m.transferModel, cmd = m.transferModel.Update(msg)
			return m, cmd
		}

		if m.view == viewAPIGateway {
			if msg.String() == "esc" && m.apiGatewayModel.state != APIGatewayStateMenu {
				return m, m.apiGatewayModel.showMenu()
			}
			if msg.String() == "esc" && m.apiGatewayModel.state == APIGatewayStateMenu {
				m.view = viewHome
				return m, nil
			}
			m.apiGatewayModel, cmd = m.apiGatewayModel.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "r": // Manual refresh
			if m.view == viewHome {
				m.cache.Delete(m.cacheKeys.Identity())
				return m, m.fetchIdentity()
			}
		case "tab":
			// No-op, header is non-interactive
		case "up":
			if m.selectedService > 0 {
				m.selectedService--
			} else {
				// Move to previous category in the same column
				col := m.getCategoryColumn(m.selectedCategory)
				categoriesInCol := m.getCategoriesInColumn(col)
				for i, catIdx := range categoriesInCol {
					if catIdx == m.selectedCategory {
						prevIdx := (i - 1 + len(categoriesInCol)) % len(categoriesInCol)
						m.selectedCategory = categoriesInCol[prevIdx]
						m.selectedService = len(m.categories[m.selectedCategory].Services) - 1
						break
					}
				}
			}
		case "down":
			if m.selectedService < len(m.categories[m.selectedCategory].Services)-1 {
				m.selectedService++
			} else {
				// Move to next category in the same column
				col := m.getCategoryColumn(m.selectedCategory)
				categoriesInCol := m.getCategoriesInColumn(col)
				for i, catIdx := range categoriesInCol {
					if catIdx == m.selectedCategory {
						nextIdx := (i + 1) % len(categoriesInCol)
						m.selectedCategory = categoriesInCol[nextIdx]
						m.selectedService = 0
						break
					}
				}
			}
		case "left":
			col := m.getCategoryColumn(m.selectedCategory)
			if col > 0 {
				newCol := col - 1
				m.moveToColumn(newCol)
			}
		case "right":
			col := m.getCategoryColumn(m.selectedCategory)
			if col < 2 {
				newCol := col + 1
				m.moveToColumn(newCol)
			}
		case "enter":
			selectedService := m.categories[m.selectedCategory].Services[m.selectedService]
			return m.handleServiceSelection(selectedService)
		}

	case ProfileSelectedMsg:
		m.selectedProfile = string(msg)
		m.profileSelector.active = false
		m.identity = nil                                     // Clear current identity
		m.cacheKeys = cache.NewKeyBuilder(m.selectedProfile) // Update cache keys for new profile
		// Reset views if profile changed
		if m.view == viewS3 {
			m.s3Model = NewS3Model(m.selectedProfile, m.styles, m.cache)
			m.s3Model.SetSize(m.width, m.height)
			return m, tea.Batch(m.s3Model.Init(), m.fetchIdentity())
		}
		if m.view == viewIAM {
			m.iamModel = NewIAMModel(m.selectedProfile, m.styles, m.cache)
			m.iamModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.iamModel.Init(), m.fetchIdentity())
		}
		if m.view == viewVPC {
			m.vpcModel = NewVPCModel(m.selectedProfile, m.styles, m.cache)
			m.vpcModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.vpcModel.Init(), m.fetchIdentity())
		}
		if m.view == viewLambda {
			m.lambdaModel = NewLambdaModel(m.selectedProfile, m.styles, m.cache)
			m.lambdaModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.lambdaModel.Init(), m.fetchIdentity())
		}
		if m.view == viewEC2 {
			m.ec2Model = NewEC2Model(m.selectedProfile, m.styles, m.cache)
			m.ec2Model.SetSize(m.width, m.height)
			return m, tea.Batch(m.ec2Model.Init(), m.fetchIdentity())
		}
		if m.view == viewRDS {
			m.rdsModel = NewRDSModel(m.selectedProfile, m.styles, m.cache)
			m.rdsModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.rdsModel.Init(), m.fetchIdentity())
		}
		if m.view == viewCW {
			m.cwModel = NewCWModel(m.selectedProfile, m.styles, m.cache)
			m.cwModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.cwModel.Init(), m.fetchIdentity())
		}
		if m.view == viewCF {
			m.cfModel = NewCFModel(m.selectedProfile, m.styles, m.cache)
			m.cfModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.cfModel.Init(), m.fetchIdentity())
		}
		if m.view == viewElastiCache {
			m.elasticacheModel = NewElastiCacheModel(m.selectedProfile, m.styles, m.cache)
			m.elasticacheModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.elasticacheModel.Init(), m.fetchIdentity())
		}
		if m.view == viewMSK {
			m.mskModel = NewMSKModel(m.selectedProfile, m.styles, m.cache)
			m.mskModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.mskModel.Init(), m.fetchIdentity())
		}
		if m.view == viewSQS {
			m.sqsModel = NewSQSModel(m.selectedProfile, m.styles, m.cache)
			m.sqsModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.sqsModel.Init(), m.fetchIdentity())
		}
		if m.view == viewSM {
			m.smModel = NewSMModel(m.selectedProfile, m.styles, m.cache)
			m.smModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.smModel.Init(), m.fetchIdentity())
		}
		if m.view == viewRoute53 {
			m.route53Model = NewRoute53Model(m.selectedProfile, m.styles, m.cache)
			m.route53Model.SetSize(m.width, m.height)
			return m, tea.Batch(m.route53Model.Init(), m.fetchIdentity())
		}
		if m.view == viewACM {
			m.acmModel = NewACMModel(m.selectedProfile, m.styles, m.cache)
			m.acmModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.acmModel.Init(), m.fetchIdentity())
		}
		if m.view == viewSNS {
			m.snsModel = NewSNSModel(m.selectedProfile, m.styles, m.cache)
			m.snsModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.snsModel.Init(), m.fetchIdentity())
		}
		if m.view == viewKMS {
			m.kmsModel = NewKMSModel(m.selectedProfile, m.styles, m.cache)
			m.kmsModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.kmsModel.Init(), m.fetchIdentity())
		}
		if m.view == viewDMS {
			m.dmsModel = NewDMSModel(m.selectedProfile, m.styles, m.cache)
			m.dmsModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.dmsModel.Init(), m.fetchIdentity())
		}
		if m.view == viewECS {
			m.ecsModel = NewECSModel(m.selectedProfile, m.styles, m.cache)
			m.ecsModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.ecsModel.Init(), m.fetchIdentity())
		}
		if m.view == viewBilling {
			m.billingModel = NewBillingModel(m.selectedProfile, m.styles, m.cache)
			m.billingModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.billingModel.Init(), m.fetchIdentity())
		}
		if m.view == viewSecurityHub {
			m.securityhubModel = NewSecurityHubModel(m.selectedProfile, m.styles, m.cache)
			m.securityhubModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.securityhubModel.Init(), m.fetchIdentity())
		}
		if m.view == viewWAF {
			region := "us-east-1"
			if m.identity != nil {
				region = m.identity.Region
			}
			m.wafModel = NewWAFModel(m.selectedProfile, m.styles, m.cache, region)
			m.wafModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.wafModel.Init(), m.fetchIdentity())
		}
		if m.view == viewECR {
			m.ecrModel = NewECRModel(m.selectedProfile, m.styles, m.cache)
			m.ecrModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.ecrModel.Init(), m.fetchIdentity())
		}
		if m.view == viewEFS {
			m.efsModel = NewEFSModel(m.selectedProfile, m.styles, m.cache)
			m.efsModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.efsModel.Init(), m.fetchIdentity())
		}
		if m.view == viewBackup {
			m.backupModel = NewBackupModel(m.selectedProfile, m.styles, m.cache)
			m.backupModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.backupModel.Init(), m.fetchIdentity())
		}
		if m.view == viewDynamoDB {
			m.dynamodbModel = NewDynamoDBModel(m.selectedProfile, m.styles, m.cache)
			m.dynamodbModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.dynamodbModel.Init(), m.fetchIdentity())
		}
		if m.view == viewTransfer {
			m.transferModel = NewTransferModel(m.selectedProfile, m.styles, m.cache)
			m.transferModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.transferModel.Init(), m.fetchIdentity())
		}
		if m.view == viewAPIGateway {
			m.apiGatewayModel = NewAPIGatewayModel(m.selectedProfile, m.styles, m.cache)
			m.apiGatewayModel.SetSize(m.width, m.height)
			return m, tea.Batch(m.apiGatewayModel.Init(), m.fetchIdentity())
		}
		return m, m.fetchIdentity()

	case S3BucketsMsg, S3ObjectsMsg, S3ErrorMsg, S3SuccessMsg:
		m.s3Model, cmd = m.s3Model.Update(msg)
		return m, cmd

	case IAMUsersMsg, IAMErrorMsg, IAMSuccessMsg:
		m.iamModel, cmd = m.iamModel.Update(msg)
		return m, cmd

	case VPCsMsg, SubnetsMsg, NatGatewaysMsg, RouteTablesMsg, VpnGatewaysMsg, VPCErrorMsg, VPCMenuMsg:
		m.vpcModel, cmd = m.vpcModel.Update(msg)
		return m, cmd

	case LambdaFunctionsMsg, LambdaErrorMsg:
		m.lambdaModel, cmd = m.lambdaModel.Update(msg)
		return m, cmd

	case InstancesMsg, SecurityGroupsMsg, VolumesMsg, TargetGroupsMsg, EC2ErrorMsg, EC2MenuMsg:
		m.ec2Model, cmd = m.ec2Model.Update(msg)
		return m, cmd

	case RDSInstancesMsg, RDSClustersMsg, RDSSnapshotsMsg, RDSSubnetGroupsMsg, RDSErrorMsg, RDSMenuMsg:
		m.rdsModel, cmd = m.rdsModel.Update(msg)
		return m, cmd

	case CWLogGroupsMsg, CWLogStreamsMsg, CWLogEventsMsg, CWErrorMsg, CWMenuMsg:
		m.cwModel, cmd = m.cwModel.Update(msg)
		return m, cmd

	case CFDistributionsMsg, CFOriginsMsg, CFBehaviorsMsg, CFInvalidationsMsg, CFPoliciesMsg, CFFunctionsMsg, CFErrorMsg, CFMenuMsg:
		m.cfModel, cmd = m.cfModel.Update(msg)
		return m, cmd

	case ReplicationGroupsMsg, CacheClustersMsg, ElastiCacheErrorMsg, ElastiCacheMenuMsg:
		m.elasticacheModel, cmd = m.elasticacheModel.Update(msg)
		return m, cmd

	case MSKClustersMsg, MSKErrorMsg:
		m.mskModel, cmd = m.mskModel.Update(msg)
		return m, cmd

	case SQSQueuesMsg, SQSErrorMsg:
		m.sqsModel, cmd = m.sqsModel.Update(msg)
		return m, cmd

	case SMSecretsMsg, SMSecretValueMsg, SMErrorMsg:
		m.smModel, cmd = m.smModel.Update(msg)
		return m, cmd

	case HostedZonesMsg, RecordSetsMsg, Route53ErrorMsg:
		m.route53Model, cmd = m.route53Model.Update(msg)
		return m, cmd

	case CertificatesMsg, ACMErrorMsg:
		m.acmModel, cmd = m.acmModel.Update(msg)
		return m, cmd

	case SNSTopicsMsg, SNSErrorMsg:
		m.snsModel, cmd = m.snsModel.Update(msg)
		return m, cmd

	case KMSKeysMsg, KMSErrorMsg:
		m.kmsModel, cmd = m.kmsModel.Update(msg)
		return m, cmd

	case DMSTasksMsg, DMSEndpointsMsg, DMSInstancesMsg, DMSErrorMsg, DMSSuccessMsg:
		m.dmsModel, cmd = m.dmsModel.Update(msg)
		return m, cmd

	case ECSClustersMsg, ECSServicesMsg, ECSTasksMsg, ECSEventsMsg, ECSTaskDefsMsg, ECSTaskDefFamiliesMsg, ECSTaskDefJSONMsg, ECSErrorMsg, ECSSuccessMsg:
		m.ecsModel, cmd = m.ecsModel.Update(msg)
		return m, cmd

	case BillingMsg, BillingErrorMsg:
		m.billingModel, cmd = m.billingModel.Update(msg)
		return m, cmd

	case SecurityHubMsg, SecurityHubErrorMsg:
		m.securityhubModel, cmd = m.securityhubModel.Update(msg)
		return m, cmd

	case WAFWebACLsMsg, WAFIPSetsMsg, WAFErrorMsg, WAFMenuMsg:
		if m.view == viewWAF {
			m.wafModel, cmd = m.wafModel.Update(msg)
			return m, cmd
		}

	case ECRReposMsg, ECRImagesMsg, ECRErrorMsg:
		if m.view == viewECR {
			m.ecrModel, cmd = m.ecrModel.Update(msg)
			return m, cmd
		}

	case EFSFileSystemsMsg, EFSMountTargetsMsg, EFSErrorMsg:
		if m.view == viewEFS {
			m.efsModel, cmd = m.efsModel.Update(msg)
			return m, cmd
		}

	case BackupPlansMsg, BackupJobsMsg, BackupErrorMsg:
		if m.view == viewBackup {
			m.backupModel, cmd = m.backupModel.Update(msg)
			return m, cmd
		}

	case DynamoTablesMsg, DynamoErrorMsg:
		if m.view == viewDynamoDB {
			m.dynamodbModel, cmd = m.dynamodbModel.Update(msg)
			return m, cmd
		}

	case TransferServersMsg, TransferUsersMsg, TransferErrorMsg:
		if m.view == viewTransfer {
			m.transferModel, cmd = m.transferModel.Update(msg)
			return m, cmd
		}

	case APIGatewayRestAPIsMsg, APIGatewayHTTPAPIsMsg, APIGatewayErrorMsg, APIGatewayMenuMsg:
		if m.view == viewAPIGateway {
			m.apiGatewayModel, cmd = m.apiGatewayModel.Update(msg)
			return m, cmd
		}

	case ECSLogGroupMsg:
		m.view = viewCW
		m.cwModel = NewCWModel(m.selectedProfile, m.styles, m.cache)
		m.cwModel.SetSize(m.width, m.height)
		m.cwModel.selectedGroup = string(msg)
		m.cwModel.state = CWStateLogStreams
		m.cwModel.originView = viewECS
		return m, m.cwModel.fetchLogStreams(string(msg))

	case SSMStartedMsg:
		c := exec.Command("aws", "ssm", "start-session", "--target", string(msg), "--profile", m.selectedProfile)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return EC2ErrorMsg(err)
			}
			return nil
		})

	case IdentityMsg:
		m.identity = msg
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// 1. DYNAMIC HEADER TITLE
	titleText := "AWS TUI"
	if m.view == viewS3 {
		titleParts := []string{"S3"}
		if m.s3Model.currentBucket != "" {
			titleParts = append(titleParts, "Buckets", m.s3Model.currentBucket)
			if m.s3Model.currentPrefix != "" {
				titleParts = append(titleParts, strings.TrimSuffix(m.s3Model.currentPrefix, "/"))
			}
		} else {
			titleParts = append(titleParts, "Buckets")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewIAM {
		titleParts := []string{"IAM", "Users"}
		if m.iamModel.state == IAMStateActions || m.iamModel.state == IAMStateConfirmDelete || m.iamModel.state == IAMStateConfirmConsoleToggle {
			titleParts = append(titleParts, m.iamModel.selectedUser.userName)
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewVPC {
		titleParts := []string{"VPC"}
		switch m.vpcModel.state {
		case VPCStateMenu:
			titleParts = append(titleParts, "Network")
		case VPCStateVPCs:
			titleParts = append(titleParts, "VPCs")
		case VPCStateSubnets:
			titleParts = append(titleParts, "Subnets")
		case VPCStateNatGateways:
			titleParts = append(titleParts, "NAT Gateways")
		case VPCStateRouteTables:
			titleParts = append(titleParts, "Route Tables")
		case VPCStateVpnGateways:
			titleParts = append(titleParts, "VPN Gateways")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewLambda {
		titleText = "Lambda / Functions"
	} else if m.view == viewEC2 {
		titleParts := []string{"EC2"}
		switch m.ec2Model.state {
		case EC2StateMenu:
			titleParts = append(titleParts, "Resources")
		case EC2StateInstances:
			titleParts = append(titleParts, "Instances")
		case EC2StateSecurityGroups:
			titleParts = append(titleParts, "Security Groups")
		case EC2StateVolumes:
			titleParts = append(titleParts, "Volumes")
		case EC2StateTargetGroups:
			titleParts = append(titleParts, "Target Groups")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewRDS {
		titleParts := []string{"RDS"}
		switch m.rdsModel.state {
		case RDSStateMenu:
			titleParts = append(titleParts, "Databases")
		case RDSStateInstances:
			titleParts = append(titleParts, "Databases")
		case RDSStateClusters:
			titleParts = append(titleParts, "Clusters")
		case RDSStateSnapshots:
			titleParts = append(titleParts, "Snapshots")
		case RDSStateSubnetGroups:
			titleParts = append(titleParts, "Subnet Groups")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewCW {
		titleParts := []string{"CloudWatch"}
		switch m.cwModel.state {
		case CWStateMenu:
			titleParts = append(titleParts, "Logs")
		case CWStateLogGroups:
			titleParts = append(titleParts, "Log Groups")
		case CWStateLogStreams:
			titleParts = append(titleParts, "Log Groups", m.cwModel.selectedGroup)
		case CWStateLogEvents:
			titleParts = append(titleParts, "Log Groups", m.cwModel.selectedGroup, m.cwModel.selectedStream)
		case CWStateLogDetail:
			titleParts = append(titleParts, "Log Groups", m.cwModel.selectedGroup, m.cwModel.selectedStream, "Detail")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewCF {
		titleParts := []string{"CloudFront"}
		switch m.cfModel.state {
		case CFStateMenu:
			// Just "CloudFront" is fine
		case CFStateDistributions:
			titleParts = append(titleParts, "Distributions")
		case CFStateDistroSubMenu:
			titleParts = append(titleParts, "Distributions", m.cfModel.selectedDistro)
		case CFStateOrigins:
			titleParts = append(titleParts, "Distributions", m.cfModel.selectedDistro, "Origins")
		case CFStateBehaviors:
			titleParts = append(titleParts, "Distributions", m.cfModel.selectedDistro, "Behaviors")
		case CFStateInvalidations:
			titleParts = append(titleParts, "Distributions", m.cfModel.selectedDistro, "Invalidations")
		case CFStatePolicies:
			titleParts = append(titleParts, "Policies")
		case CFStateFunctions:
			titleParts = append(titleParts, "Functions")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewElastiCache {
		titleParts := []string{"ElastiCache"}
		switch m.elasticacheModel.state {
		case ElastiCacheStateMenu:
			titleParts = append(titleParts, "Resources")
		case ElastiCacheStateReplicationGroups:
			titleParts = append(titleParts, "Replication Groups")
		case ElastiCacheStateCacheClusters:
			titleParts = append(titleParts, "Cache Clusters")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewMSK {
		titleText = "MSK / Clusters"
	} else if m.view == viewSQS {
		titleText = "SQS / Queues"
	} else if m.view == viewSM {
		titleParts := []string{"Secrets Manager"}
		switch m.smModel.state {
		case SMStateSecrets:
			titleParts = append(titleParts, "Secrets")
		case SMStateValue:
			titleParts = append(titleParts, "Secrets", m.smModel.selectedSecret)
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewRoute53 {
		titleParts := []string{"Route 53"}
		switch m.route53Model.state {
		case Route53StateZones:
			titleParts = append(titleParts, "Zones")
		case Route53StateRecords:
			titleParts = append(titleParts, "Zones", m.route53Model.selectedZoneName, "Records")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewACM {
		titleText = "ACM / Certificates"
	} else if m.view == viewSNS {
		titleText = "SNS / Topics"
	} else if m.view == viewKMS {
		titleText = "KMS / Keys"
	} else if m.view == viewDMS {
		titleParts := []string{"DMS"}
		switch m.dmsModel.state {
		case DMSStateMenu:
			// Just DMS
		case DMSStateTasks:
			titleParts = append(titleParts, "Tasks")
		case DMSStateEndpoints:
			titleParts = append(titleParts, "Endpoints")
		case DMSStateInstances:
			titleParts = append(titleParts, "Instances")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewECS {
		titleParts := []string{"ECS"}
		if m.ecsModel.selectedCluster != "" {
			titleParts = append(titleParts, m.ecsModel.selectedCluster)
			if m.ecsModel.selectedService != "" {
				titleParts = append(titleParts, m.ecsModel.selectedService)
				switch m.ecsModel.state {
				case ECSStateTasks:
					titleParts = append(titleParts, "Tasks")
				case ECSStateEvents:
					titleParts = append(titleParts, "Events")
				}
			} else {
				titleParts = append(titleParts, "Services")
			}
		} else if m.ecsModel.state == ECSStateTaskDefFamilies {
			titleParts = append(titleParts, "Task Definitions")
		} else if m.ecsModel.state == ECSStateTaskDefRevisions {
			titleParts = append(titleParts, "Task Definitions", m.ecsModel.selectedTaskDefFamily)
		} else {
			titleParts = append(titleParts, "Resources")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewBilling {
		titleText = "Billing / Costs"
	} else if m.view == viewSecurityHub {
		titleText = "Security Hub / Findings"
	} else if m.view == viewWAF {
		titleParts := []string{"WAFv2"}
		switch m.wafModel.state {
		case WAFStateMenu:
			titleParts = append(titleParts, "Resources")
		case WAFStateWebACLs:
			titleParts = append(titleParts, string(m.wafModel.scope), "Web ACLs")
		case WAFStateIPSets:
			titleParts = append(titleParts, string(m.wafModel.scope), "IP Sets")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewECR {
		titleParts := []string{"ECR"}
		if m.ecrModel.currentRepository != "" {
			titleParts = append(titleParts, "Repositories", m.ecrModel.currentRepository)
			if m.ecrModel.state == ECRStateImages {
				titleParts = append(titleParts, "Images")
			}
		} else {
			titleParts = append(titleParts, "Repositories")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewEFS {
		titleParts := []string{"EFS"}
		if m.efsModel.currentFileSystem != "" {
			titleParts = append(titleParts, "File Systems", m.efsModel.currentFileSystem)
			if m.efsModel.state == EFSStateMountTargets {
				titleParts = append(titleParts, "Mount Targets")
			}
		} else {
			titleParts = append(titleParts, "File Systems")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewBackup {
		titleParts := []string{"AWS Backup"}
		switch m.backupModel.state {
		case BackupStatePlans:
			titleParts = append(titleParts, "Plans")
		case BackupStateJobs:
			titleParts = append(titleParts, "Jobs")
		}
		titleText = strings.Join(titleParts, " / ")
	} else if m.view == viewDynamoDB {
		titleText = "DynamoDB / Tables"
		} else if m.view == viewTransfer {
			titleParts := []string{"AWS Transfer"}
			if m.transferModel.currentServer != "" {
				titleParts = append(titleParts, "Servers", m.transferModel.currentServer)
				if m.transferModel.state == TransferStateUsers {
					titleParts = append(titleParts, "Users")
				}
			} else {
				titleParts = append(titleParts, "Servers")
			}
			titleText = strings.Join(titleParts, " / ")
		} else if m.view == viewAPIGateway {
			titleParts := []string{"API Gateway"}
			switch m.apiGatewayModel.state {
			case APIGatewayStateMenu:
				titleParts = append(titleParts, "Resources")
			case APIGatewayStateRestAPIs:
				titleParts = append(titleParts, "REST APIs")
			case APIGatewayStateHTTPAPIs:
				titleParts = append(titleParts, "HTTP APIs")
			}
			titleText = strings.Join(titleParts, " / ")
		}
	currentViewTitle := m.styles.ViewTitle.Render(titleText)

	// Profile Section
	profileLabel := m.styles.StatusKey.Render("profile: ")
	profileValue := m.selectedProfile
	profileText := profileLabel + m.styles.Profile.Render(profileValue)

	// Session Info (Account & Region)
	var sessionInfo string
	if m.identity != nil {
		accInfo := lipgloss.NewStyle().Foreground(m.styles.Snow).Render(m.identity.Account)
		if m.identity.Alias != "" {
			accInfo += " " + m.styles.StatusMuted.Render("("+m.identity.Alias+")")
		}

		region := m.identity.Region
		if region == "" {
			region = "unknown"
		}
		regionInfo := lipgloss.NewStyle().Foreground(m.styles.Snow).Render(region)

		sessionInfo = lipgloss.JoinHorizontal(lipgloss.Center,
			m.styles.StatusMuted.Render(" | "),
			m.styles.StatusKey.Render("account: "), accInfo,
			m.styles.StatusMuted.Render(" | "),
			m.styles.StatusKey.Render("region: "), regionInfo,
		)
	} else {
		sessionInfo = m.styles.StatusMuted.Render(" | ") + m.styles.StatusMuted.Render("loading session...")
	}

	headerContent := lipgloss.JoinHorizontal(lipgloss.Center,
		currentViewTitle,
		m.styles.StatusMuted.Render(" | "),
		profileText,
		sessionInfo,
	)

	// Center the content inside the header box
	headerWidth := m.width - AppWidthOffset
	centeredHeaderContent := lipgloss.NewStyle().
		Width(headerWidth - 2). // -2 for borders
		Align(lipgloss.Center).
		Render(headerContent)

	header := m.styles.Header.Width(headerWidth).Render(centeredHeaderContent)

	// 2. FOOTER HINTS
	footerHints := []string{
		m.styles.StatusKey.Render("↑↓←→") + " " + m.styles.StatusMuted.Render("Navigate"),
		m.styles.StatusKey.Render("/") + " " + m.styles.StatusMuted.Render("Filter"),
		m.styles.StatusKey.Render("Enter") + " " + m.styles.StatusMuted.Render("Select"),
	}

	if m.view != viewHome {
		footerHints = append(footerHints, m.styles.StatusKey.Render("esc")+" "+m.styles.StatusMuted.Render("Back"))
	}

	footerHints = append(footerHints,
		m.styles.StatusKey.Render("p")+" "+m.styles.StatusMuted.Render("Profile"),
		m.styles.StatusKey.Render("r")+" "+m.styles.StatusMuted.Render("Refresh"),
	)

	// Context-specific hints
	if m.view == viewS3 {
		if m.s3Model.state == S3StateBuckets {
			footerHints = append(footerHints, m.styles.StatusKey.Render("n")+" "+m.styles.StatusMuted.Render("New Bucket"))
		} else if m.s3Model.state == S3StateObjects {
			footerHints = append(footerHints,
				m.styles.StatusKey.Render("n")+" "+m.styles.StatusMuted.Render("New Folder"),
				m.styles.StatusKey.Render("u")+" "+m.styles.StatusMuted.Render("Upload"),
				m.styles.StatusKey.Render("e")+" "+m.styles.StatusMuted.Render("Edit"),
			)
		}
		if m.s3Model.state == S3StateBuckets || m.s3Model.state == S3StateObjects {
			footerHints = append(footerHints, m.styles.StatusKey.Render("d")+" "+m.styles.StatusMuted.Render("Delete"))
		}
	} else if m.view == viewIAM {
		if m.iamModel.state == IAMStateUsers {
			footerHints = append(footerHints,
				m.styles.StatusKey.Render("n")+" "+m.styles.StatusMuted.Render("New User"),
				m.styles.StatusKey.Render("d")+" "+m.styles.StatusMuted.Render("Delete"),
			)
		}
	} else if m.view == viewDMS {
		if m.dmsModel.state == DMSStateTasks {
			footerHints = append(footerHints, m.styles.StatusKey.Render("o")+" "+m.styles.StatusMuted.Render("Options"))
		}
	} else if m.view == viewECS {
		if m.ecsModel.state == ECSStateTasks || m.ecsModel.state == ECSStateServices {
			footerHints = append(footerHints, m.styles.StatusKey.Render("o")+" "+m.styles.StatusMuted.Render("Options"))
		}
	} else if m.view == viewEC2 {
		if m.ec2Model.state == EC2StateInstances {
			footerHints = append(footerHints, m.styles.StatusKey.Render("o")+" "+m.styles.StatusMuted.Render("Options"))
		}
	} else if m.view == viewWAF {
		if m.wafModel.state != WAFStateMenu {
			footerHints = append(footerHints, m.styles.StatusKey.Render("backspace")+" "+m.styles.StatusMuted.Render("Back to Menu"))
		}
	}

	footerHints = append(footerHints, m.styles.StatusKey.Render("q")+" "+m.styles.StatusMuted.Render("Quit"))
	internalFooter := strings.Join(footerHints, m.styles.StatusMuted.Render(" • "))

	// 3. MAIN BOX CONTENT
	var boxContent string
	if m.profileSelector.active {
		popup := m.styles.Popup.Width(38).Render(
			m.profileSelector.View(),
		)
		w, h := GetMainContainerSize(m.width, m.height)
		boxContent = lipgloss.Place(w, h-AppInternalFooterHeight-2, lipgloss.Center, lipgloss.Center, popup)
	} else {
		switch m.view {
		case viewS3:
			boxContent = m.s3Model.View()
		case viewIAM:
			boxContent = m.iamModel.View()
		case viewVPC:
			boxContent = m.vpcModel.View()
		case viewLambda:
			boxContent = m.lambdaModel.View()
		case viewEC2:
			boxContent = m.ec2Model.View()
		case viewRDS:
			boxContent = m.rdsModel.View()
		case viewCW:
			boxContent = m.cwModel.View()
		case viewCF:
			boxContent = m.cfModel.View()
		case viewElastiCache:
			boxContent = m.elasticacheModel.View()
		case viewMSK:
			boxContent = m.mskModel.View()
		case viewSQS:
			boxContent = m.sqsModel.View()
		case viewSM:
			boxContent = m.smModel.View()
		case viewRoute53:
			boxContent = m.route53Model.View()
		case viewACM:
			boxContent = m.acmModel.View()
		case viewSNS:
			boxContent = m.snsModel.View()
		case viewKMS:
			boxContent = m.kmsModel.View()
		case viewDMS:
			boxContent = m.dmsModel.View()
		case viewECS:
			boxContent = m.ecsModel.View()
		case viewBilling:
			boxContent = m.billingModel.View()
		case viewSecurityHub:
			boxContent = m.securityhubModel.View()
		case viewWAF:
			boxContent = m.wafModel.View()
		case viewECR:
			boxContent = m.ecrModel.View()
		case viewEFS:
			boxContent = m.efsModel.View()
		case viewBackup:
			boxContent = m.backupModel.View()
		case viewDynamoDB:
			boxContent = m.dynamodbModel.View()
		case viewTransfer:
			boxContent = m.transferModel.View()
		case viewAPIGateway:
			boxContent = m.apiGatewayModel.View()
		default:
			// Home View
			logo := `
	  ___   ____    __    ____   _______.   .___________. __    __   __
     /   \  \   \  /  \  /   /  /       |   |           ||  |  |  | |  |
    /  ^  \  \   \/    \/   /  |   (----` + "`    ---|  |----`|  |  |  | |  | " + `
   /  /_\  \  \            /    \   \           |  |     |  |  |  | |  |
  /  _____  \  \    /\    / .----)   |          |  |     |  ` + "`" + `--'  | |  |
 /__/     \__\  \__/  \__/  |_______/           |__|      \______/  |__|`
			logoStyle := lipgloss.NewStyle().Foreground(m.styles.Primary).Bold(true)

			subtitle := lipgloss.NewStyle().
				Foreground(m.styles.Muted).
				Italic(true).
				MarginBottom(1).
				MarginTop(1).
				Render("Manage your AWS infrastructure without leaving your shell.")

			// Services Menu
			var menuBox string
			if m.searching {
				var sb strings.Builder
				sb.WriteString(m.searchInput.View() + "\n")

				// Show max 10 items
				displayItems := m.filteredServices
				if len(displayItems) > 10 {
					displayItems = displayItems[:10]
				}

				if len(displayItems) == 0 {
					sb.WriteString(m.styles.StatusMuted.Render("No services found.") + "\n")
				} else {
					for i, service := range displayItems {
						icon := featureIcons[service]
						if icon == "" {
							icon = "• "
						}
						if i == m.selectedFiltered {
							sb.WriteString(m.styles.SelectedMenuItem.Render("➜ "+icon+service) + "\n")
						} else {
							sb.WriteString(m.styles.MenuItem.Render("  "+icon+service) + "\n")
						}
					}
				}

				// Fixed size box: width 60, height 14 (1 line input + 1 blank + 10 items + 2 border)
				// Content height: 1 (input) + 1 (blank) + 10 (items) = 12 lines
				menuBox = m.styles.MenuContainer.Copy().
					Border(lipgloss.RoundedBorder()).
					Padding(1, 2).
					Width(60).
					Height(14).
					Render(sb.String())
			} else {
				renderCategory := func(catIdx int) string {
					cat := m.categories[catIdx]
					var sb strings.Builder
					sb.WriteString(lipgloss.NewStyle().
						Foreground(m.styles.Primary).
						Bold(true).
						Underline(true).
						MarginBottom(1).
						Render(strings.ToUpper(cat.Name)) + "\n")

					for i, service := range cat.Services {
						icon := featureIcons[service]
						if icon == "" {
							icon = "• "
						}

						if catIdx == m.selectedCategory && i == m.selectedService && m.focus == focusContent {
							sb.WriteString(m.styles.SelectedMenuItem.Render("➜ "+icon+service) + "\n")
						} else {
							sb.WriteString(m.styles.MenuItem.Render("  "+icon+service) + "\n")
						}
					}
					return sb.String()
				}

				col0 := lipgloss.JoinVertical(lipgloss.Left,
					renderCategory(0),
					"\n",
					renderCategory(1),
				)
				col1 := lipgloss.JoinVertical(lipgloss.Left,
					renderCategory(2),
					"\n",
					renderCategory(3),
					"\n",
					renderCategory(5),
				)
				col2 := lipgloss.JoinVertical(lipgloss.Left,
					renderCategory(4),
					"\n",
					renderCategory(6),
				)

				columns := lipgloss.JoinHorizontal(lipgloss.Top,
					lipgloss.NewStyle().Width(40).Render(col0),
					lipgloss.NewStyle().Width(40).Render(col1),
					lipgloss.NewStyle().Width(40).Render(col2),
				)

				menuBox = m.styles.MenuContainer.Copy().
					Border(lipgloss.RoundedBorder()).
					Padding(1, 2).
					Render(columns)
			}

			homeView := lipgloss.JoinVertical(lipgloss.Center,
				logoStyle.Render(logo),
				subtitle,
				menuBox,
			)

			headerHeight := lipgloss.Height(header)
			boxContent = lipgloss.Place(m.width-InnerContentWidthOffset, m.height-headerHeight-7, lipgloss.Center, lipgloss.Center, homeView)
		}
	}

	mainBox := RenderBoxedContainer(m.styles, boxContent, internalFooter, m.width, m.height)

	// 4. JOIN EVERYTHING VERTICALLY
	finalView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainBox,
	)

	return finalView
}

func (m Model) getCategoryColumn(categoryIdx int) int {
	switch categoryIdx {
	case 0, 1:
		return 0
	case 2, 3, 5:
		return 1
	case 4, 6:
		return 2
	default:
		return 0
	}
}

func (m Model) getCategoriesInColumn(col int) []int {
	switch col {
	case 0:
		return []int{0, 1}
	case 1:
		return []int{2, 3, 5}
	case 2:
		return []int{4, 6}
	default:
		return []int{}
	}
}

func (m *Model) moveToColumn(newCol int) {
	oldCol := m.getCategoryColumn(m.selectedCategory)
	categoriesInOldCol := m.getCategoriesInColumn(oldCol)

	oldRank := 0
	for i, catIdx := range categoriesInOldCol {
		if catIdx == m.selectedCategory {
			oldRank = i
			break
		}
	}

	categoriesInNewCol := m.getCategoriesInColumn(newCol)
	if len(categoriesInNewCol) > 0 {
		newRank := oldRank
		if newRank >= len(categoriesInNewCol) {
			newRank = len(categoriesInNewCol) - 1
		}
		m.selectedCategory = categoriesInNewCol[newRank]
		if m.selectedService >= len(m.categories[m.selectedCategory].Services) {
			m.selectedService = len(m.categories[m.selectedCategory].Services) - 1
		}
	}
}

func (m Model) renderMainContainer(content string, footer string) string {
	return RenderBoxedContainer(m.styles, content, footer, m.width, m.height)
}

func (m Model) isInputFocused() bool {
	if m.searching {
		return true
	}
	if m.view == viewS3 && m.s3Model.state == S3StateInput {
		return true
	}
	if m.view == viewIAM && m.iamModel.state == IAMStateInput {
		return true
	}
	return false
}

func (m *Model) handleServiceSelection(selectedService string) (tea.Model, tea.Cmd) {
	handlers := getServiceHandlers()
	if handler, ok := handlers[selectedService]; ok {
		return handler(m)
	}
	return *m, nil
}

func (m *Model) updateFilter() {
	query := m.searchInput.Value()
	allServices := []string{}
	for _, cat := range m.categories {
		allServices = append(allServices, cat.Services...)
	}

	if query == "" {
		m.filteredServices = allServices
	} else {
		matches := fuzzy.Find(query, allServices)
		m.filteredServices = make([]string, len(matches))
		for i, match := range matches {
			m.filteredServices[i] = match.Str
		}
	}

	// Limit to max 10 items
	maxItems := 10
	if len(m.filteredServices) > maxItems {
		m.filteredServices = m.filteredServices[:maxItems]
	}

	if m.selectedFiltered >= len(m.filteredServices) {
		m.selectedFiltered = len(m.filteredServices) - 1
		if m.selectedFiltered < 0 {
			m.selectedFiltered = 0
		}
	}
	if m.selectedFiltered < 0 && len(m.filteredServices) > 0 {
		m.selectedFiltered = 0
	}
}
