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
			"ElastiCache (Todo)",
			"MSK (Todo)",
			"SQS Queues (Todo)",
			"Secrets Manager (Todo)",
			"Route 53 Zones (Todo)",
			"ACM Certificates (Todo)",
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
		titleText = "S3 Buckets"
	} else if m.view == viewIAM {
		if m.iamModel.state == IAMStateActions {
			titleText = "IAM Actions"
		} else {
			titleText = "IAM Users"
		}
	} else if m.view == viewVPC {
		switch m.vpcModel.state {
		case VPCStateMenu:
			titleText = "VPC Network"
		case VPCStateVPCs:
			titleText = "VPCs"
		case VPCStateSubnets:
			titleText = "Subnets"
		case VPCStateNatGateways:
			titleText = "NAT Gateways"
		case VPCStateRouteTables:
			titleText = "Route Tables"
		case VPCStateVpnGateways:
			titleText = "VPN Gateways"
		}
	} else if m.view == viewLambda {
		titleText = "Lambda Functions"
	} else if m.view == viewEC2 {
		switch m.ec2Model.state {
		case EC2StateMenu:
			titleText = "EC2 Resources"
		case EC2StateInstances:
			titleText = "Instances"
		case EC2StateSecurityGroups:
			titleText = "Security Groups"
		case EC2StateVolumes:
			titleText = "Volumes"
		case EC2StateTargetGroups:
			titleText = "Target Groups"
		}
	} else if m.view == viewRDS {
		switch m.rdsModel.state {
		case RDSStateMenu:
			titleText = "RDS Databases"
		case RDSStateInstances:
			titleText = "Databases"
		case RDSStateClusters:
			titleText = "Clusters"
		case RDSStateSnapshots:
			titleText = "Snapshots"
		case RDSStateSubnetGroups:
			titleText = "Subnet Groups"
		}
	} else if m.view == viewCW {
		switch m.cwModel.state {
		case CWStateMenu:
			titleText = "CloudWatch Logs"
		case CWStateLogGroups:
			titleText = "Log Groups"
		case CWStateLogStreams:
			titleText = "Log Streams"
		case CWStateLogEvents:
			titleText = "Log Events"
		case CWStateLogDetail:
			titleText = "Log Detail"
		}
	} else if m.view == viewCF {
		switch m.cfModel.state {
		case CFStateMenu:
			titleText = "CloudFront"
		case CFStateDistributions:
			titleText = "Distributions"
		case CFStateDistroSubMenu:
			titleText = "Distribution Details"
		case CFStateOrigins:
			titleText = "Origins"
		case CFStateBehaviors:
			titleText = "Behaviors"
		case CFStateInvalidations:
			titleText = "Invalidations"
		case CFStatePolicies:
			titleText = "Policies"
		case CFStateFunctions:
			titleText = "Functions"
		}
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
				"SQS Queues (Todo)":         "󰒔 ",
				"ElastiCache (Todo)":        "󰓡 ",
				"Secrets Manager (Todo)":    "󰌆 ",
				"ACM Certificates (Todo)":   "󰔕 ",
				"Route 53 Zones (Todo)":     "󰇧 ",
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
