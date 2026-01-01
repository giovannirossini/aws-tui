package ui

import (
	"context"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

type Model struct {
	profiles        []string
	selectedProfile string
	profileSelector ProfileSelector
	styles          Styles
	focus           focus
	view            viewState
	s3Model         S3Model
	iamModel        IAMModel
	vpcModel        VPCModel
	lambdaModel     LambdaModel
	ec2Model        EC2Model
	rdsModel        RDSModel
	cwModel         CWModel
	cfModel         CFModel
	elasticacheModel ElastiCacheModel
	mskModel        MSKModel
	sqsModel        SQSModel
	smModel         SMModel
	route53Model    Route53Model
	acmModel        ACMModel
	features        []string
	selectedFeature int
	width           int
	height          int
	ready           bool
	identity        *aws.IdentityInfo
	cache           *cache.Cache
	cacheKeys       *cache.KeyBuilder
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

	return Model{
		profiles:        profiles,
		selectedProfile: selected,
		profileSelector: ps,
		styles:          styles,
		focus:           focusContent,
		view:            viewHome,
		features: []string{
			"S3 Buckets",
			"IAM Users",
			"VPC Network",
			"Lambda Functions",
			"EC2 Resources",
			"RDS Databases",
			"CloudWatch Logs",
			"CloudFront Distros",
			"ElastiCache",
			"MSK",
			"SQS Queues",
			"Secrets Manager",
			"Route 53 Zones",
			"ACM Certificates",
			"SNS Topics (Todo)",
			"KMS Keys (Todo)",
		},
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(selected),
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
		m.ready = true

	case tea.KeyMsg:
		if m.profileSelector.active {
			m.profileSelector, cmd = m.profileSelector.Update(msg)
			return m, cmd
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

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r": // Manual refresh
			if m.view == viewHome {
				m.cache.Delete(m.cacheKeys.Identity())
				return m, m.fetchIdentity()
			}
		case "p", "P":
			m.profileSelector.active = true
			m.profileSelector.list.FilterInput.Focus()
			return m, nil
		case "tab":
			// No-op, header is non-interactive
		case "up":
			if m.selectedFeature > 0 {
				m.selectedFeature--
			}
		case "down":
			if m.selectedFeature < len(m.features)-1 {
				m.selectedFeature++
			}
		case "enter":
			if m.features[m.selectedFeature] == "S3 Buckets" {
				m.view = viewS3
				m.s3Model = NewS3Model(m.selectedProfile, m.styles, m.cache)
				m.s3Model.SetSize(m.width, m.height)
				return m, m.s3Model.Init()
			}
			if m.features[m.selectedFeature] == "IAM Users" {
				m.view = viewIAM
				m.iamModel = NewIAMModel(m.selectedProfile, m.styles, m.cache)
				m.iamModel.SetSize(m.width, m.height)
				return m, m.iamModel.Init()
			}
			if m.features[m.selectedFeature] == "VPC Network" {
				m.view = viewVPC
				m.vpcModel = NewVPCModel(m.selectedProfile, m.styles, m.cache)
				m.vpcModel.SetSize(m.width, m.height)
				return m, m.vpcModel.Init()
			}
			if m.features[m.selectedFeature] == "Lambda Functions" {
				m.view = viewLambda
				m.lambdaModel = NewLambdaModel(m.selectedProfile, m.styles, m.cache)
				m.lambdaModel.SetSize(m.width, m.height)
				return m, m.lambdaModel.Init()
			}
			if m.features[m.selectedFeature] == "EC2 Resources" {
				m.view = viewEC2
				m.ec2Model = NewEC2Model(m.selectedProfile, m.styles, m.cache)
				m.ec2Model.SetSize(m.width, m.height)
				return m, m.ec2Model.Init()
			}
			if m.features[m.selectedFeature] == "RDS Databases" {
				m.view = viewRDS
				m.rdsModel = NewRDSModel(m.selectedProfile, m.styles, m.cache)
				m.rdsModel.SetSize(m.width, m.height)
				return m, m.rdsModel.Init()
			}
			if m.features[m.selectedFeature] == "CloudWatch Logs" {
				m.view = viewCW
				m.cwModel = NewCWModel(m.selectedProfile, m.styles, m.cache)
				m.cwModel.SetSize(m.width, m.height)
				return m, m.cwModel.Init()
			}
			if m.features[m.selectedFeature] == "CloudFront Distros" {
				m.view = viewCF
				m.cfModel = NewCFModel(m.selectedProfile, m.styles, m.cache)
				m.cfModel.SetSize(m.width, m.height)
				return m, m.cfModel.Init()
			}
			if m.features[m.selectedFeature] == "ElastiCache" {
				m.view = viewElastiCache
				m.elasticacheModel = NewElastiCacheModel(m.selectedProfile, m.styles, m.cache)
				m.elasticacheModel.SetSize(m.width, m.height)
				return m, m.elasticacheModel.Init()
			}
			if m.features[m.selectedFeature] == "MSK" {
				m.view = viewMSK
				m.mskModel = NewMSKModel(m.selectedProfile, m.styles, m.cache)
				m.mskModel.SetSize(m.width, m.height)
				return m, m.mskModel.Init()
			}
			if m.features[m.selectedFeature] == "SQS Queues" {
				m.view = viewSQS
				m.sqsModel = NewSQSModel(m.selectedProfile, m.styles, m.cache)
				m.sqsModel.SetSize(m.width, m.height)
				return m, m.sqsModel.Init()
			}
			if m.features[m.selectedFeature] == "Secrets Manager" {
				m.view = viewSM
				m.smModel = NewSMModel(m.selectedProfile, m.styles, m.cache)
				m.smModel.SetSize(m.width, m.height)
				return m, m.smModel.Init()
			}
			if m.features[m.selectedFeature] == "Route 53 Zones" {
				m.view = viewRoute53
				m.route53Model = NewRoute53Model(m.selectedProfile, m.styles, m.cache)
				m.route53Model.SetSize(m.width, m.height)
				return m, m.route53Model.Init()
			}
			if m.features[m.selectedFeature] == "ACM Certificates" {
				m.view = viewACM
				m.acmModel = NewACMModel(m.selectedProfile, m.styles, m.cache)
				m.acmModel.SetSize(m.width, m.height)
				return m, m.acmModel.Init()
			}
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
		m.styles.StatusKey.Render("↑↓") + " " + m.styles.StatusMuted.Render("Navigate"),
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
				Render("Manage your AWS infrastructure without leaving your shell.")

			// Services Menu
			var menuContent strings.Builder
			menuContent.WriteString(lipgloss.NewStyle().Foreground(m.styles.Primary).Bold(true).Render(" AVAILABLE SERVICES ") + "\n\n")

			featureIcons := map[string]string{
				"S3 Buckets":                "󱐖 ",
				"IAM Users":                 " ",
				"VPC Network":               "󰛳 ",
				"Lambda Functions":          "󰘧 ",
				"EC2 Resources":             " ",
				"RDS Databases":            "󰆼 ",
				"CloudWatch Logs":           "󱖉 ",
				"CloudFront Distros":        "󰇄 ",
				"ElastiCache":               "󰓡 ",
				"MSK":                       "󰒔 ",
				"SQS Queues":                "󰒔 ",
				"Secrets Manager":           "󰌆 ",
				"Route 53 Zones":            "󰇧 ",
				"ACM Certificates":          "󰔕 ",
				"SNS Topics (Todo)":         "󰰓 ",
				"KMS Keys (Todo)":           "󰌆 ",
			}

			for i, feature := range m.features {
				icon := featureIcons[feature]
				if icon == "" {
					icon = "• "
				}

				if i == m.selectedFeature && m.focus == focusContent {
					menuContent.WriteString(m.styles.SelectedMenuItem.Render("➜ "+icon+feature) + "\n")
				} else {
					menuContent.WriteString(m.styles.MenuItem.Render("  "+icon+feature) + "\n")
				}
			}

			menuBox := m.styles.MenuContainer.Copy().
				Border(lipgloss.RoundedBorder()).
				Padding(1, 2).
				Width(60).
				Render(menuContent.String())

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

func (m Model) renderMainContainer(content string, footer string) string {
	return RenderBoxedContainer(m.styles, content, footer, m.width, m.height)
}
