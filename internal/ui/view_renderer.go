package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHeader generates the dynamic header title based on current view
func (m Model) renderHeader() string {
	titleText := m.getViewTitle()

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

	return m.styles.Header.Width(headerWidth).Render(centeredHeaderContent)
}

// getViewTitle returns the title text for the current view
func (m Model) getViewTitle() string {
	switch m.view {
	case viewS3:
		titleParts := []string{"S3"}
		if m.s3Model.currentBucket != "" {
			titleParts = append(titleParts, "Buckets", m.s3Model.currentBucket)
			if m.s3Model.currentPrefix != "" {
				titleParts = append(titleParts, strings.TrimSuffix(m.s3Model.currentPrefix, "/"))
			}
		} else {
			titleParts = append(titleParts, "Buckets")
		}
		return strings.Join(titleParts, " / ")
	case viewIAM:
		titleParts := []string{"IAM", "Users"}
		if m.iamModel.state == IAMStateActions || m.iamModel.state == IAMStateConfirmDelete || m.iamModel.state == IAMStateConfirmConsoleToggle {
			titleParts = append(titleParts, m.iamModel.selectedUser.userName)
		}
		return strings.Join(titleParts, " / ")
	case viewVPC:
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
		return strings.Join(titleParts, " / ")
	case viewLambda:
		return "Lambda / Functions"
	case viewEC2:
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
		return strings.Join(titleParts, " / ")
	case viewRDS:
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
		return strings.Join(titleParts, " / ")
	case viewCW:
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
		return strings.Join(titleParts, " / ")
	case viewCF:
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
		return strings.Join(titleParts, " / ")
	case viewElastiCache:
		titleParts := []string{"ElastiCache"}
		switch m.elasticacheModel.state {
		case ElastiCacheStateMenu:
			titleParts = append(titleParts, "Resources")
		case ElastiCacheStateReplicationGroups:
			titleParts = append(titleParts, "Replication Groups")
		case ElastiCacheStateCacheClusters:
			titleParts = append(titleParts, "Cache Clusters")
		}
		return strings.Join(titleParts, " / ")
	case viewMSK:
		return "MSK / Clusters"
	case viewSQS:
		return "SQS / Queues"
	case viewSM:
		titleParts := []string{"Secrets Manager"}
		switch m.smModel.state {
		case SMStateSecrets:
			titleParts = append(titleParts, "Secrets")
		case SMStateValue:
			titleParts = append(titleParts, "Secrets", m.smModel.selectedSecret)
		}
		return strings.Join(titleParts, " / ")
	case viewRoute53:
		titleParts := []string{"Route 53"}
		switch m.route53Model.state {
		case Route53StateZones:
			titleParts = append(titleParts, "Zones")
		case Route53StateRecords:
			titleParts = append(titleParts, "Zones", m.route53Model.selectedZoneName, "Records")
		}
		return strings.Join(titleParts, " / ")
	case viewACM:
		return "ACM / Certificates"
	case viewSNS:
		return "SNS / Topics"
	case viewKMS:
		return "KMS / Keys"
	case viewDMS:
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
		return strings.Join(titleParts, " / ")
	case viewECS:
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
		return strings.Join(titleParts, " / ")
	case viewBilling:
		return "Billing / Costs"
	case viewSecurityHub:
		return "Security Hub / Findings"
	case viewWAF:
		titleParts := []string{"WAFv2"}
		switch m.wafModel.state {
		case WAFStateMenu:
			titleParts = append(titleParts, "Resources")
		case WAFStateWebACLs:
			titleParts = append(titleParts, string(m.wafModel.scope), "Web ACLs")
		case WAFStateIPSets:
			titleParts = append(titleParts, string(m.wafModel.scope), "IP Sets")
		}
		return strings.Join(titleParts, " / ")
	case viewECR:
		titleParts := []string{"ECR"}
		if m.ecrModel.currentRepository != "" {
			titleParts = append(titleParts, "Repositories", m.ecrModel.currentRepository)
			if m.ecrModel.state == ECRStateImages {
				titleParts = append(titleParts, "Images")
			}
		} else {
			titleParts = append(titleParts, "Repositories")
		}
		return strings.Join(titleParts, " / ")
	case viewEFS:
		titleParts := []string{"EFS"}
		if m.efsModel.currentFileSystem != "" {
			titleParts = append(titleParts, "File Systems", m.efsModel.currentFileSystem)
			if m.efsModel.state == EFSStateMountTargets {
				titleParts = append(titleParts, "Mount Targets")
			}
		} else {
			titleParts = append(titleParts, "File Systems")
		}
		return strings.Join(titleParts, " / ")
	case viewBackup:
		titleParts := []string{"AWS Backup"}
		switch m.backupModel.state {
		case BackupStatePlans:
			titleParts = append(titleParts, "Plans")
		case BackupStateJobs:
			titleParts = append(titleParts, "Jobs")
		}
		return strings.Join(titleParts, " / ")
	case viewDynamoDB:
		return "DynamoDB / Tables"
	case viewTransfer:
		titleParts := []string{"AWS Transfer"}
		if m.transferModel.currentServer != "" {
			titleParts = append(titleParts, "Servers", m.transferModel.currentServer)
			if m.transferModel.state == TransferStateUsers {
				titleParts = append(titleParts, "Users")
			}
		} else {
			titleParts = append(titleParts, "Servers")
		}
		return strings.Join(titleParts, " / ")
	case viewAPIGateway:
		titleParts := []string{"API Gateway"}
		switch m.apiGatewayModel.state {
		case APIGatewayStateMenu:
			titleParts = append(titleParts, "Resources")
		case APIGatewayStateRestAPIs:
			titleParts = append(titleParts, "REST APIs")
		case APIGatewayStateHTTPAPIs:
			titleParts = append(titleParts, "HTTP APIs")
		}
		return strings.Join(titleParts, " / ")
	default:
		return "AWS TUI"
	}
}

// renderFooter generates footer hints based on current view and context
func (m Model) renderFooter() string {
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
	m.addContextSpecificHints(&footerHints)

	footerHints = append(footerHints, m.styles.StatusKey.Render("q")+" "+m.styles.StatusMuted.Render("Quit"))
	return strings.Join(footerHints, m.styles.StatusMuted.Render(" • "))
}

// addContextSpecificHints adds view-specific footer hints
func (m Model) addContextSpecificHints(footerHints *[]string) {
	switch m.view {
	case viewS3:
		if m.s3Model.state == S3StateBuckets {
			*footerHints = append(*footerHints, m.styles.StatusKey.Render("n")+" "+m.styles.StatusMuted.Render("New Bucket"))
		} else if m.s3Model.state == S3StateObjects {
			*footerHints = append(*footerHints,
				m.styles.StatusKey.Render("n")+" "+m.styles.StatusMuted.Render("New Folder"),
				m.styles.StatusKey.Render("u")+" "+m.styles.StatusMuted.Render("Upload"),
				m.styles.StatusKey.Render("e")+" "+m.styles.StatusMuted.Render("Edit"),
			)
		}
		if m.s3Model.state == S3StateBuckets || m.s3Model.state == S3StateObjects {
			*footerHints = append(*footerHints, m.styles.StatusKey.Render("d")+" "+m.styles.StatusMuted.Render("Delete"))
		}
	case viewIAM:
		if m.iamModel.state == IAMStateUsers {
			*footerHints = append(*footerHints,
				m.styles.StatusKey.Render("n")+" "+m.styles.StatusMuted.Render("New User"),
				m.styles.StatusKey.Render("d")+" "+m.styles.StatusMuted.Render("Delete"),
			)
		}
	case viewDMS:
		if m.dmsModel.state == DMSStateTasks {
			*footerHints = append(*footerHints, m.styles.StatusKey.Render("o")+" "+m.styles.StatusMuted.Render("Options"))
		}
	case viewECS:
		if m.ecsModel.state == ECSStateTasks || m.ecsModel.state == ECSStateServices {
			*footerHints = append(*footerHints, m.styles.StatusKey.Render("o")+" "+m.styles.StatusMuted.Render("Options"))
		}
	case viewEC2:
		if m.ec2Model.state == EC2StateInstances {
			*footerHints = append(*footerHints, m.styles.StatusKey.Render("o")+" "+m.styles.StatusMuted.Render("Options"))
		}
	case viewWAF:
		if m.wafModel.state != WAFStateMenu {
			*footerHints = append(*footerHints, m.styles.StatusKey.Render("backspace")+" "+m.styles.StatusMuted.Render("Back to Menu"))
		}
	}
}

// renderMainContent renders the main content area based on current view
func (m Model) renderMainContent() string {
	if m.profileSelector.active {
		popup := m.styles.Popup.Width(38).Render(
			m.profileSelector.View(),
		)
		w, h := GetMainContainerSize(m.width, m.height)
		return lipgloss.Place(w, h-AppInternalFooterHeight-2, lipgloss.Center, lipgloss.Center, popup)
	}

	switch m.view {
	case viewS3:
		return m.s3Model.View()
	case viewIAM:
		return m.iamModel.View()
	case viewVPC:
		return m.vpcModel.View()
	case viewLambda:
		return m.lambdaModel.View()
	case viewEC2:
		return m.ec2Model.View()
	case viewRDS:
		return m.rdsModel.View()
	case viewCW:
		return m.cwModel.View()
	case viewCF:
		return m.cfModel.View()
	case viewElastiCache:
		return m.elasticacheModel.View()
	case viewMSK:
		return m.mskModel.View()
	case viewSQS:
		return m.sqsModel.View()
	case viewSM:
		return m.smModel.View()
	case viewRoute53:
		return m.route53Model.View()
	case viewACM:
		return m.acmModel.View()
	case viewSNS:
		return m.snsModel.View()
	case viewKMS:
		return m.kmsModel.View()
	case viewDMS:
		return m.dmsModel.View()
	case viewECS:
		return m.ecsModel.View()
	case viewBilling:
		return m.billingModel.View()
	case viewSecurityHub:
		return m.securityhubModel.View()
	case viewWAF:
		return m.wafModel.View()
	case viewECR:
		return m.ecrModel.View()
	case viewEFS:
		return m.efsModel.View()
	case viewBackup:
		return m.backupModel.View()
	case viewDynamoDB:
		return m.dynamodbModel.View()
	case viewTransfer:
		return m.transferModel.View()
	case viewAPIGateway:
		return m.apiGatewayModel.View()
	default:
		return m.renderHomeView()
	}
}

// renderHomeView renders the home view with service categories
func (m Model) renderHomeView() string {
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
		menuBox = m.renderSearchMenu()
	} else {
		menuBox = m.renderServiceCategories()
	}

	homeView := lipgloss.JoinVertical(lipgloss.Center,
		logoStyle.Render(logo),
		subtitle,
		menuBox,
	)

	header := m.renderHeader()
	headerHeight := lipgloss.Height(header)
	return lipgloss.Place(m.width-InnerContentWidthOffset, m.height-headerHeight-7, lipgloss.Center, lipgloss.Center, homeView)
}

// renderSearchMenu renders the search menu when searching
func (m Model) renderSearchMenu() string {
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
	return m.styles.MenuContainer.Copy().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(60).
		Height(14).
		Render(sb.String())
}

// renderServiceCategories renders the service categories in columns
func (m Model) renderServiceCategories() string {
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

	return m.styles.MenuContainer.Copy().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Render(columns)
}
