package ui

import (
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

// handleWindowSize handles window size updates for all views
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.profileSelector.SetSize(m.width, m.height)

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.view {
	case viewS3:
		m.s3Model.SetSize(m.width, m.height)
		m.s3Model, cmd = m.s3Model.Update(msg)
		cmds = append(cmds, cmd)
	case viewIAM:
		m.iamModel.SetSize(m.width, m.height)
		m.iamModel, cmd = m.iamModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewVPC:
		m.vpcModel.SetSize(m.width, m.height)
		m.vpcModel, cmd = m.vpcModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewLambda:
		m.lambdaModel.SetSize(m.width, m.height)
		m.lambdaModel, cmd = m.lambdaModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewEC2:
		m.ec2Model.SetSize(m.width, m.height)
		m.ec2Model, cmd = m.ec2Model.Update(msg)
		cmds = append(cmds, cmd)
	case viewRDS:
		m.rdsModel.SetSize(m.width, m.height)
		m.rdsModel, cmd = m.rdsModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewCW:
		m.cwModel.SetSize(m.width, m.height)
		m.cwModel, cmd = m.cwModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewCF:
		m.cfModel.SetSize(m.width, m.height)
		m.cfModel, cmd = m.cfModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewElastiCache:
		m.elasticacheModel.SetSize(m.width, m.height)
		m.elasticacheModel, cmd = m.elasticacheModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewMSK:
		m.mskModel.SetSize(m.width, m.height)
		m.mskModel, cmd = m.mskModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewSQS:
		m.sqsModel.SetSize(m.width, m.height)
		m.sqsModel, cmd = m.sqsModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewSM:
		m.smModel.SetSize(m.width, m.height)
		m.smModel, cmd = m.smModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewRoute53:
		m.route53Model.SetSize(m.width, m.height)
		m.route53Model, cmd = m.route53Model.Update(msg)
		cmds = append(cmds, cmd)
	case viewACM:
		m.acmModel.SetSize(m.width, m.height)
		m.acmModel, cmd = m.acmModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewSNS:
		m.snsModel.SetSize(m.width, m.height)
		m.snsModel, cmd = m.snsModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewKMS:
		m.kmsModel.SetSize(m.width, m.height)
		m.kmsModel, cmd = m.kmsModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewDMS:
		m.dmsModel.SetSize(m.width, m.height)
		m.dmsModel, cmd = m.dmsModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewECS:
		m.ecsModel.SetSize(m.width, m.height)
		m.ecsModel, cmd = m.ecsModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewBilling:
		m.billingModel.SetSize(m.width, m.height)
		m.billingModel, cmd = m.billingModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewSecurityHub:
		m.securityhubModel.SetSize(m.width, m.height)
		m.securityhubModel, cmd = m.securityhubModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewWAF:
		m.wafModel.SetSize(m.width, m.height)
		m.wafModel, cmd = m.wafModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewECR:
		m.ecrModel.SetSize(m.width, m.height)
		m.ecrModel, cmd = m.ecrModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewEFS:
		m.efsModel.SetSize(m.width, m.height)
		m.efsModel, cmd = m.efsModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewBackup:
		m.backupModel.SetSize(m.width, m.height)
		m.backupModel, cmd = m.backupModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewDynamoDB:
		m.dynamodbModel.SetSize(m.width, m.height)
		m.dynamodbModel, cmd = m.dynamodbModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewTransfer:
		m.transferModel.SetSize(m.width, m.height)
		m.transferModel, cmd = m.transferModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewAPIGateway:
		m.apiGatewayModel.SetSize(m.width, m.height)
		m.apiGatewayModel, cmd = m.apiGatewayModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.ready = true
	return *m, tea.Batch(cmds...)
}

// handleKeyPress routes key presses to appropriate handlers
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.profileSelector.active {
		var cmd tea.Cmd
		m.profileSelector, cmd = m.profileSelector.Update(msg)
		return *m, cmd
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
				return *m, textinput.Blink
			}
		case "p", "P":
			m.profileSelector.active = true
			m.profileSelector.list.FilterInput.Focus()
			return *m, nil
		case "q", "ctrl+c":
			return *m, tea.Quit
		}
	} else {
		// Even in input mode, ctrl+c should quit
		if msg.String() == "ctrl+c" {
			return *m, tea.Quit
		}

		if m.searching {
			return m.handleSearchInput(msg)
		}
	}

	// Route to view-specific handlers
	if cmd := m.handleViewKeyPress(msg); cmd != nil {
		return *m, cmd
	}

	// Handle home view navigation
	if m.view == viewHome {
		return m.handleHomeNavigation(msg)
	}

	return *m, nil
}

// handleSearchInput handles search input interactions
func (m *Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searching = false
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		return *m, nil
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
	return *m, tiCmd
}

// handleViewKeyPress handles key presses for specific views
func (m *Model) handleViewKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch m.view {
	case viewS3:
		return m.handleS3KeyPress(msg)
	case viewIAM:
		return m.handleIAMKeyPress(msg)
	case viewVPC:
		return m.handleVPCKeyPress(msg)
	case viewLambda:
		return m.handleLambdaKeyPress(msg)
	case viewEC2:
		return m.handleEC2KeyPress(msg)
	case viewRDS:
		return m.handleRDSKeyPress(msg)
	case viewCW:
		return m.handleCWKeyPress(msg)
	case viewCF:
		return m.handleCFKeyPress(msg)
	case viewElastiCache:
		return m.handleElastiCacheKeyPress(msg)
	case viewMSK:
		return m.handleMSKKeyPress(msg)
	case viewSQS:
		return m.handleSQSKeyPress(msg)
	case viewSM:
		return m.handleSMKeyPress(msg)
	case viewRoute53:
		return m.handleRoute53KeyPress(msg)
	case viewACM:
		return m.handleACMKeyPress(msg)
	case viewSNS:
		return m.handleSNSKeyPress(msg)
	case viewKMS:
		return m.handleKMSKeyPress(msg)
	case viewDMS:
		return m.handleDMSKeyPress(msg)
	case viewECS:
		return m.handleECSKeyPress(msg)
	case viewBilling:
		return m.handleBillingKeyPress(msg)
	case viewSecurityHub:
		return m.handleSecurityHubKeyPress(msg)
	case viewWAF:
		return m.handleWAFKeyPress(msg)
	case viewECR:
		return m.handleECRKeyPress(msg)
	case viewEFS:
		return m.handleEFSKeyPress(msg)
	case viewBackup:
		return m.handleBackupKeyPress(msg)
	case viewDynamoDB:
		return m.handleDynamoDBKeyPress(msg)
	case viewTransfer:
		return m.handleTransferKeyPress(msg)
	case viewAPIGateway:
		return m.handleAPIGatewayKeyPress(msg)
	}
	return nil
}

// Individual view key press handlers
func (m *Model) handleS3KeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.s3Model.state == S3StateBuckets {
		m.view = viewHome
		return nil
	}
	// Special handling for edit which requires suspension
	if msg.String() == "e" && m.s3Model.state == S3StateObjects {
		if item, ok := m.s3Model.list.SelectedItem().(s3Item); ok && !item.isFolder && !item.isBucket {
			return tea.ExecProcess(m.s3Model.getEditCommand(item.key), func(err error) tea.Msg {
				if err != nil {
					return S3ErrorMsg(err)
				}
				return m.s3Model.uploadEditedFile(item.key)
			})
		}
	}
	var cmd tea.Cmd
	m.s3Model, cmd = m.s3Model.Update(msg)
	return cmd
}

func (m *Model) handleIAMKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.iamModel.state == IAMStateUsers {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.iamModel, cmd = m.iamModel.Update(msg)
	return cmd
}

func (m *Model) handleVPCKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.vpcModel.state == VPCStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.vpcModel, cmd = m.vpcModel.Update(msg)
	return cmd
}

func (m *Model) handleLambdaKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.lambdaModel, cmd = m.lambdaModel.Update(msg)
	return cmd
}

func (m *Model) handleEC2KeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.ec2Model.state == EC2StateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.ec2Model, cmd = m.ec2Model.Update(msg)
	return cmd
}

func (m *Model) handleRDSKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.rdsModel.state == RDSStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.rdsModel, cmd = m.rdsModel.Update(msg)
	return cmd
}

func (m *Model) handleCWKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.cwModel.state == CWStateLogStreams && m.cwModel.originView == viewECS {
		m.view = viewECS
		return nil
	}
	if msg.String() == "esc" && m.cwModel.state == CWStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.cwModel, cmd = m.cwModel.Update(msg)
	return cmd
}

func (m *Model) handleCFKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.cfModel.state == CFStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.cfModel, cmd = m.cfModel.Update(msg)
	return cmd
}

func (m *Model) handleElastiCacheKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.elasticacheModel.state == ElastiCacheStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.elasticacheModel, cmd = m.elasticacheModel.Update(msg)
	return cmd
}

func (m *Model) handleMSKKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.mskModel, cmd = m.mskModel.Update(msg)
	return cmd
}

func (m *Model) handleSQSKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.sqsModel, cmd = m.sqsModel.Update(msg)
	return cmd
}

func (m *Model) handleSMKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.smModel.state == SMStateSecrets {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.smModel, cmd = m.smModel.Update(msg)
	return cmd
}

func (m *Model) handleRoute53KeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.route53Model.state == Route53StateZones {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.route53Model, cmd = m.route53Model.Update(msg)
	return cmd
}

func (m *Model) handleACMKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.acmModel, cmd = m.acmModel.Update(msg)
	return cmd
}

func (m *Model) handleSNSKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.snsModel, cmd = m.snsModel.Update(msg)
	return cmd
}

func (m *Model) handleKMSKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.kmsModel, cmd = m.kmsModel.Update(msg)
	return cmd
}

func (m *Model) handleDMSKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.dmsModel.state == DMSStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.dmsModel, cmd = m.dmsModel.Update(msg)
	return cmd
}

func (m *Model) handleECSKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.ecsModel.state == ECSStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.ecsModel, cmd = m.ecsModel.Update(msg)
	return cmd
}

func (m *Model) handleBillingKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.billingModel, cmd = m.billingModel.Update(msg)
	return cmd
}

func (m *Model) handleSecurityHubKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.securityhubModel, cmd = m.securityhubModel.Update(msg)
	return cmd
}

func (m *Model) handleWAFKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.wafModel.state == WAFStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.wafModel, cmd = m.wafModel.Update(msg)
	return cmd
}

func (m *Model) handleECRKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.ecrModel.state == ECRStateRepositories {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.ecrModel, cmd = m.ecrModel.Update(msg)
	return cmd
}

func (m *Model) handleEFSKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.efsModel.state == EFSStateFileSystems {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.efsModel, cmd = m.efsModel.Update(msg)
	return cmd
}

func (m *Model) handleBackupKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.backupModel.state == BackupStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.backupModel, cmd = m.backupModel.Update(msg)
	return cmd
}

func (m *Model) handleDynamoDBKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.dynamodbModel, cmd = m.dynamodbModel.Update(msg)
	return cmd
}

func (m *Model) handleTransferKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.transferModel.state == TransferStateServers {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.transferModel, cmd = m.transferModel.Update(msg)
	return cmd
}

func (m *Model) handleAPIGatewayKeyPress(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "esc" && m.apiGatewayModel.state != APIGatewayStateMenu {
		return m.apiGatewayModel.showMenu()
	}
	if msg.String() == "esc" && m.apiGatewayModel.state == APIGatewayStateMenu {
		m.view = viewHome
		return nil
	}
	var cmd tea.Cmd
	m.apiGatewayModel, cmd = m.apiGatewayModel.Update(msg)
	return cmd
}

// handleHomeNavigation handles navigation keys in the home view
func (m *Model) handleHomeNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r": // Manual refresh
		m.cache.Delete(m.cacheKeys.Identity())
		return *m, m.fetchIdentity()
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
	return *m, nil
}

// handleProfileChange handles profile switching and resets the current view
func (m *Model) handleProfileChange(profile string) (tea.Model, tea.Cmd) {
	m.selectedProfile = profile
	m.profileSelector.active = false
	m.identity = nil
	m.cacheKeys = cache.NewKeyBuilder(m.selectedProfile)

	// Reset current view with new profile
	switch m.view {
	case viewS3:
		m.s3Model = NewS3Model(m.selectedProfile, m.styles, m.cache)
		m.s3Model.SetSize(m.width, m.height)
		return *m, tea.Batch(m.s3Model.Init(), m.fetchIdentity())
	case viewIAM:
		m.iamModel = NewIAMModel(m.selectedProfile, m.styles, m.cache)
		m.iamModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.iamModel.Init(), m.fetchIdentity())
	case viewVPC:
		m.vpcModel = NewVPCModel(m.selectedProfile, m.styles, m.cache)
		m.vpcModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.vpcModel.Init(), m.fetchIdentity())
	case viewLambda:
		m.lambdaModel = NewLambdaModel(m.selectedProfile, m.styles, m.cache)
		m.lambdaModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.lambdaModel.Init(), m.fetchIdentity())
	case viewEC2:
		m.ec2Model = NewEC2Model(m.selectedProfile, m.styles, m.cache)
		m.ec2Model.SetSize(m.width, m.height)
		return *m, tea.Batch(m.ec2Model.Init(), m.fetchIdentity())
	case viewRDS:
		m.rdsModel = NewRDSModel(m.selectedProfile, m.styles, m.cache)
		m.rdsModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.rdsModel.Init(), m.fetchIdentity())
	case viewCW:
		m.cwModel = NewCWModel(m.selectedProfile, m.styles, m.cache)
		m.cwModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.cwModel.Init(), m.fetchIdentity())
	case viewCF:
		m.cfModel = NewCFModel(m.selectedProfile, m.styles, m.cache)
		m.cfModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.cfModel.Init(), m.fetchIdentity())
	case viewElastiCache:
		m.elasticacheModel = NewElastiCacheModel(m.selectedProfile, m.styles, m.cache)
		m.elasticacheModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.elasticacheModel.Init(), m.fetchIdentity())
	case viewMSK:
		m.mskModel = NewMSKModel(m.selectedProfile, m.styles, m.cache)
		m.mskModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.mskModel.Init(), m.fetchIdentity())
	case viewSQS:
		m.sqsModel = NewSQSModel(m.selectedProfile, m.styles, m.cache)
		m.sqsModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.sqsModel.Init(), m.fetchIdentity())
	case viewSM:
		m.smModel = NewSMModel(m.selectedProfile, m.styles, m.cache)
		m.smModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.smModel.Init(), m.fetchIdentity())
	case viewRoute53:
		m.route53Model = NewRoute53Model(m.selectedProfile, m.styles, m.cache)
		m.route53Model.SetSize(m.width, m.height)
		return *m, tea.Batch(m.route53Model.Init(), m.fetchIdentity())
	case viewACM:
		m.acmModel = NewACMModel(m.selectedProfile, m.styles, m.cache)
		m.acmModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.acmModel.Init(), m.fetchIdentity())
	case viewSNS:
		m.snsModel = NewSNSModel(m.selectedProfile, m.styles, m.cache)
		m.snsModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.snsModel.Init(), m.fetchIdentity())
	case viewKMS:
		m.kmsModel = NewKMSModel(m.selectedProfile, m.styles, m.cache)
		m.kmsModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.kmsModel.Init(), m.fetchIdentity())
	case viewDMS:
		m.dmsModel = NewDMSModel(m.selectedProfile, m.styles, m.cache)
		m.dmsModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.dmsModel.Init(), m.fetchIdentity())
	case viewECS:
		m.ecsModel = NewECSModel(m.selectedProfile, m.styles, m.cache)
		m.ecsModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.ecsModel.Init(), m.fetchIdentity())
	case viewBilling:
		m.billingModel = NewBillingModel(m.selectedProfile, m.styles, m.cache)
		m.billingModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.billingModel.Init(), m.fetchIdentity())
	case viewSecurityHub:
		m.securityhubModel = NewSecurityHubModel(m.selectedProfile, m.styles, m.cache)
		m.securityhubModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.securityhubModel.Init(), m.fetchIdentity())
	case viewWAF:
		region := "us-east-1"
		if m.identity != nil {
			region = m.identity.Region
		}
		m.wafModel = NewWAFModel(m.selectedProfile, m.styles, m.cache, region)
		m.wafModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.wafModel.Init(), m.fetchIdentity())
	case viewECR:
		m.ecrModel = NewECRModel(m.selectedProfile, m.styles, m.cache)
		m.ecrModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.ecrModel.Init(), m.fetchIdentity())
	case viewEFS:
		m.efsModel = NewEFSModel(m.selectedProfile, m.styles, m.cache)
		m.efsModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.efsModel.Init(), m.fetchIdentity())
	case viewBackup:
		m.backupModel = NewBackupModel(m.selectedProfile, m.styles, m.cache)
		m.backupModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.backupModel.Init(), m.fetchIdentity())
	case viewDynamoDB:
		m.dynamodbModel = NewDynamoDBModel(m.selectedProfile, m.styles, m.cache)
		m.dynamodbModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.dynamodbModel.Init(), m.fetchIdentity())
	case viewTransfer:
		m.transferModel = NewTransferModel(m.selectedProfile, m.styles, m.cache)
		m.transferModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.transferModel.Init(), m.fetchIdentity())
	case viewAPIGateway:
		m.apiGatewayModel = NewAPIGatewayModel(m.selectedProfile, m.styles, m.cache)
		m.apiGatewayModel.SetSize(m.width, m.height)
		return *m, tea.Batch(m.apiGatewayModel.Init(), m.fetchIdentity())
	}
	return *m, m.fetchIdentity()
}

// handleViewMessages delegates service-specific messages to their models
func (m *Model) handleViewMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case S3BucketsMsg, S3ObjectsMsg, S3ErrorMsg, S3SuccessMsg:
		m.s3Model, cmd = m.s3Model.Update(msg)
		return *m, cmd

	case IAMUsersMsg, IAMErrorMsg, IAMSuccessMsg:
		m.iamModel, cmd = m.iamModel.Update(msg)
		return *m, cmd

	case VPCsMsg, SubnetsMsg, NatGatewaysMsg, RouteTablesMsg, VpnGatewaysMsg, VPCErrorMsg, VPCMenuMsg:
		m.vpcModel, cmd = m.vpcModel.Update(msg)
		return *m, cmd

	case LambdaFunctionsMsg, LambdaErrorMsg:
		m.lambdaModel, cmd = m.lambdaModel.Update(msg)
		return *m, cmd

	case InstancesMsg, SecurityGroupsMsg, VolumesMsg, TargetGroupsMsg, EC2ErrorMsg, EC2MenuMsg:
		m.ec2Model, cmd = m.ec2Model.Update(msg)
		return *m, cmd

	case RDSInstancesMsg, RDSClustersMsg, RDSSnapshotsMsg, RDSSubnetGroupsMsg, RDSErrorMsg, RDSMenuMsg:
		m.rdsModel, cmd = m.rdsModel.Update(msg)
		return *m, cmd

	case CWLogGroupsMsg, CWLogStreamsMsg, CWLogEventsMsg, CWErrorMsg, CWMenuMsg:
		m.cwModel, cmd = m.cwModel.Update(msg)
		return *m, cmd

	case CFDistributionsMsg, CFOriginsMsg, CFBehaviorsMsg, CFInvalidationsMsg, CFPoliciesMsg, CFFunctionsMsg, CFErrorMsg, CFMenuMsg:
		m.cfModel, cmd = m.cfModel.Update(msg)
		return *m, cmd

	case ReplicationGroupsMsg, CacheClustersMsg, ElastiCacheErrorMsg, ElastiCacheMenuMsg:
		m.elasticacheModel, cmd = m.elasticacheModel.Update(msg)
		return *m, cmd

	case MSKClustersMsg, MSKErrorMsg:
		m.mskModel, cmd = m.mskModel.Update(msg)
		return *m, cmd

	case SQSQueuesMsg, SQSErrorMsg:
		m.sqsModel, cmd = m.sqsModel.Update(msg)
		return *m, cmd

	case SMSecretsMsg, SMSecretValueMsg, SMErrorMsg:
		m.smModel, cmd = m.smModel.Update(msg)
		return *m, cmd

	case HostedZonesMsg, RecordSetsMsg, Route53ErrorMsg:
		m.route53Model, cmd = m.route53Model.Update(msg)
		return *m, cmd

	case CertificatesMsg, ACMErrorMsg:
		m.acmModel, cmd = m.acmModel.Update(msg)
		return *m, cmd

	case SNSTopicsMsg, SNSErrorMsg:
		m.snsModel, cmd = m.snsModel.Update(msg)
		return *m, cmd

	case KMSKeysMsg, KMSErrorMsg:
		m.kmsModel, cmd = m.kmsModel.Update(msg)
		return *m, cmd

	case DMSTasksMsg, DMSEndpointsMsg, DMSInstancesMsg, DMSErrorMsg, DMSSuccessMsg:
		m.dmsModel, cmd = m.dmsModel.Update(msg)
		return *m, cmd

	case ECSClustersMsg, ECSServicesMsg, ECSTasksMsg, ECSEventsMsg, ECSTaskDefsMsg, ECSTaskDefFamiliesMsg, ECSTaskDefJSONMsg, ECSErrorMsg, ECSSuccessMsg:
		m.ecsModel, cmd = m.ecsModel.Update(msg)
		return *m, cmd

	case BillingMsg, BillingErrorMsg:
		m.billingModel, cmd = m.billingModel.Update(msg)
		return *m, cmd

	case SecurityHubMsg, SecurityHubErrorMsg:
		m.securityhubModel, cmd = m.securityhubModel.Update(msg)
		return *m, cmd

	case WAFWebACLsMsg, WAFIPSetsMsg, WAFErrorMsg, WAFMenuMsg:
		if m.view == viewWAF {
			m.wafModel, cmd = m.wafModel.Update(msg)
			return *m, cmd
		}

	case ECRReposMsg, ECRImagesMsg, ECRErrorMsg:
		if m.view == viewECR {
			m.ecrModel, cmd = m.ecrModel.Update(msg)
			return *m, cmd
		}

	case EFSFileSystemsMsg, EFSMountTargetsMsg, EFSErrorMsg:
		if m.view == viewEFS {
			m.efsModel, cmd = m.efsModel.Update(msg)
			return *m, cmd
		}

	case BackupPlansMsg, BackupJobsMsg, BackupErrorMsg:
		if m.view == viewBackup {
			m.backupModel, cmd = m.backupModel.Update(msg)
			return *m, cmd
		}

	case DynamoTablesMsg, DynamoErrorMsg:
		if m.view == viewDynamoDB {
			m.dynamodbModel, cmd = m.dynamodbModel.Update(msg)
			return *m, cmd
		}

	case TransferServersMsg, TransferUsersMsg, TransferErrorMsg:
		if m.view == viewTransfer {
			m.transferModel, cmd = m.transferModel.Update(msg)
			return *m, cmd
		}

	case APIGatewayRestAPIsMsg, APIGatewayHTTPAPIsMsg, APIGatewayErrorMsg, APIGatewayMenuMsg:
		if m.view == viewAPIGateway {
			m.apiGatewayModel, cmd = m.apiGatewayModel.Update(msg)
			return *m, cmd
		}

	case ECSLogGroupMsg:
		m.view = viewCW
		m.cwModel = NewCWModel(m.selectedProfile, m.styles, m.cache)
		m.cwModel.SetSize(m.width, m.height)
		m.cwModel.selectedGroup = string(msg)
		m.cwModel.state = CWStateLogStreams
		m.cwModel.originView = viewECS
		return *m, m.cwModel.fetchLogStreams(string(msg))

	case SSMStartedMsg:
		c := exec.Command("aws", "ssm", "start-session", "--target", string(msg), "--profile", m.selectedProfile)
		return *m, tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return EC2ErrorMsg(err)
			}
			return nil
		})

	case IdentityMsg:
		m.identity = msg
		return *m, nil
	}

	return *m, nil
}
