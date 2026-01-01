package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
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

func (m Model) Init() tea.Cmd {
	return m.fetchIdentity()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case ProfileSelectedMsg:
		return m.handleProfileChange(string(msg))
	default:
		return m.handleViewMessages(msg)
	}
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	boxContent := m.renderMainContent()

	mainBox := RenderBoxedContainer(m.styles, boxContent, footer, m.width, m.height)

	// Join everything vertically
	finalView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainBox,
	)

	return finalView
}
