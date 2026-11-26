package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ryequ/redis-cockpit/internal/command"
	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/ops"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
	"github.com/Ryequ/redis-cockpit/internal/result"
	"github.com/Ryequ/redis-cockpit/internal/security"
	"github.com/Ryequ/redis-cockpit/internal/stream"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectedMsg:
		return m.handleConnected(msg)
	case profileTestedMsg:
		return m.handleProfileTested(msg)
	case commandExecutedMsg:
		return m.handleCommandExecuted(msg)
	case keysLoadedMsg:
		return m.handleKeysLoaded(msg)
	case valueLoadedMsg:
		return m.handleValueLoaded(msg)
	case keysDeletedMsg:
		return m.handleKeysDeleted(msg)
	case opsLoadedMsg:
		return m.handleOpsLoaded(msg)
	case subscribedMsg:
		return m.handleSubscribed(msg)
	case publishedMsg:
		return m.handlePublished(msg)
	case monitorStartedMsg:
		return m.handleMonitorStarted(msg)
	case streamMessageMsg:
		return m.handleStreamMessage(msg)
	case monitorEventMsg:
		return m.handleMonitorEvent(msg)
	case streamClosedMsg:
		if msg.gen != m.generation {
			return m, nil
		}
		m.setStatus("Stream closed")
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if isQuitKey(msg) {
		m.closeClient()
		return m, tea.Quit
	}

	if m.showHelpModal {
		if msg.Type == tea.KeyEsc || msg.Type == tea.KeyEnter {
			m.showHelpModal = false
		}
		return m, nil
	}

	if m.pendingDelete != nil {
		return m.handleDeleteConfirmation(msg)
	}

	if m.pendingCommand != nil {
		return m.handleConfirmation(msg)
	}

	if m.pendingProfileDelete != "" {
		return m.handleProfileDeleteConfirmation(msg)
	}

	if m.profileForm != nil {
		return m.handleProfileFormKey(msg)
	}

	if isCycleForward(msg) {
		m.nextView(1)
		return m, nil
	}
	if isCycleBackward(msg) {
		m.nextView(-1)
		return m, nil
	}

	if m.activeView == ViewWorkbench {
		return m.handleWorkbenchKey(msg)
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		if view, ok := viewFromNumber(msg.Runes[0]); ok {
			m.setView(view)
			return m, nil
		}
		if msg.Runes[0] == '?' {
			m.showHelpModal = true
			return m, nil
		}
	}

	if m.activeView == ViewKeys {
		return m.handleKeysViewKey(msg)
	}

	if m.activeView == ViewOps {
		return m.handleOpsKey(msg)
	}

	if m.activeView == ViewStreams {
		return m.handleStreamsKey(msg)
	}

	if m.activeView == ViewConnections {
		return m.handleConnectionsKey(msg)
	}

	return m, nil
}

func (m Model) handleConnectionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case 'n':
			form := newProfileForm(redis.ConnectionProfile{}, profileCreate)
			m.profileForm = &form
			m.setStatus("New connection profile")
			return m, nil
		case 'k':
			m.selectPreviousProfile()
			return m, nil
		case 'j':
			m.selectNextProfile()
			return m, nil
		case 'e':
			if len(m.cfg.Profiles) > 0 {
				form := newProfileForm(m.cfg.Profiles[m.selectedProfile], profileEdit)
				m.profileForm = &form
				m.setStatus("Editing connection profile")
			}
			return m, nil
		case 'd':
			if len(m.cfg.Profiles) > 0 {
				m.pendingProfileDelete = m.cfg.Profiles[m.selectedProfile].Name
				m.setStatus("Delete profile? press y to confirm, n to cancel")
			}
			return m, nil
		case 't':
			if len(m.cfg.Profiles) > 0 {
				profile := m.cfg.Profiles[m.selectedProfile]
				m.setStatus("Testing " + profile.Name + "...")
				return m, testConnectionCmd(profile, m.generation)
			}
			return m, nil
		}
	}

	if len(m.cfg.Profiles) == 0 {
		return m, nil
	}

	switch msg.Type {
	case tea.KeyUp:
		m.selectPreviousProfile()
		return m, nil
	case tea.KeyDown:
		m.selectNextProfile()
		return m, nil
	case tea.KeyEnter:
		profile := m.cfg.Profiles[m.selectedProfile]
		m.setStatus("Connecting to " + profile.Name + "...")
		return m, connectCmd(profile, m.generation)
	}
	return m, nil
}

func (m Model) handleProfileFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.profileForm
	if form == nil {
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.profileForm = nil
		m.setStatus("Profile edit cancelled")
		return m, nil
	case tea.KeyEnter:
		profile, err := form.profile()
		if err != nil {
			m.setStatus("Invalid profile: " + err.Error())
			return m, nil
		}
		for _, existing := range m.cfg.Profiles {
			if existing.Name == profile.Name && existing.Name != form.originalName {
				m.setStatus("Profile already exists: " + profile.Name)
				return m, nil
			}
		}

		beforeProfiles := append([]redis.ConnectionProfile(nil), m.cfg.Profiles...)
		beforeActive := m.cfg.ActiveProfile
		if form.mode == profileEdit && form.originalName != profile.Name {
			m.cfg.DeleteProfile(form.originalName)
		}
		if err := m.cfg.UpsertProfile(profile); err != nil {
			m.cfg.Profiles = beforeProfiles
			m.cfg.ActiveProfile = beforeActive
			m.setStatus("Invalid profile: " + err.Error())
			return m, nil
		}
		m.cfg.ActiveProfile = profile.Name
		if err := m.saveConfig(); err != nil {
			m.cfg.Profiles = beforeProfiles
			m.cfg.ActiveProfile = beforeActive
			m.setStatus("Save failed: " + err.Error())
			return m, nil
		}
		for i, existing := range m.cfg.Profiles {
			if existing.Name == profile.Name {
				m.selectedProfile = i
				break
			}
		}
		m.cfg.ActiveProfile = profile.Name
		m.profileForm = nil
		m.setStatus("Profile saved: " + profile.Name)
		return m, nil
	case tea.KeyTab:
		form.nextFocus()
		return m, nil
	case tea.KeyShiftTab:
		form.prevFocus()
		return m, nil
	case tea.KeySpace:
		if form.focus >= profileFieldCount {
			form.toggleBool()
			return m, nil
		}
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	}

	if form.focus < profileFieldCount {
		updated, cmd := form.inputs[form.focus].Update(msg)
		form.inputs[form.focus] = updated
		return m, cmd
	}
	return m, nil
}

func (m Model) handleProfileDeleteConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	name := m.pendingProfileDelete
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case 'y', 'Y':
			m.pendingProfileDelete = ""
			beforeProfiles := append([]redis.ConnectionProfile(nil), m.cfg.Profiles...)
			beforeActive := m.cfg.ActiveProfile
			beforeSelected := m.selectedProfile
			if !m.cfg.DeleteProfile(name) {
				m.setStatus("Profile not found: " + name)
				return m, nil
			}
			if m.cfg.ActiveProfile == "" && len(m.cfg.Profiles) > 0 {
				m.cfg.ActiveProfile = m.cfg.Profiles[0].Name
			}
			if m.selectedProfile >= len(m.cfg.Profiles) {
				m.selectedProfile = len(m.cfg.Profiles) - 1
			}
			if m.selectedProfile < 0 {
				m.selectedProfile = 0
			}
			if err := m.saveConfig(); err != nil {
				m.cfg.Profiles = beforeProfiles
				m.cfg.ActiveProfile = beforeActive
				m.selectedProfile = beforeSelected
				m.setStatus("Save failed: " + err.Error())
				return m, nil
			}
			m.setStatus("Profile deleted: " + name)
			return m, nil
		case 'n', 'N':
			m.pendingProfileDelete = ""
			m.setStatus("Delete cancelled")
			return m, nil
		}
	}
	if msg.Type == tea.KeyEsc {
		m.pendingProfileDelete = ""
		m.setStatus("Delete cancelled")
	}
	return m, nil
}

func (m Model) handleProfileTested(msg profileTestedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	name := msg.name
	if name == "" {
		name = "connection"
	}
	if msg.err != nil {
		m.setStatus("Test failed for " + name + ": " + msg.err.Error())
		return m, nil
	}
	m.setStatus("Test OK: " + name)
	return m, nil
}

func (m *Model) saveConfig() error {
	path := m.configPath
	if path == "" {
		defaultPath, err := config.DefaultPath()
		if err != nil {
			return err
		}
		path = defaultPath
	}
	return config.Save(path, m.cfg)
}

func (m Model) handleWorkbenchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		line := strings.TrimSpace(m.commandInput.Value())
		if line == "" {
			return m, nil
		}
		return m.executeLine(line)
	}
	if msg.Type == tea.KeyEsc {
		m.commandInput.SetValue("")
		m.setStatus("Command input cleared")
		return m, nil
	}

	var cmd tea.Cmd
	m.commandInput, cmd = m.commandInput.Update(msg)
	return m, cmd
}

func (m Model) handleConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case 'y', 'Y':
			argv := m.pendingCommand
			m.pendingCommand = nil
			m.commandInput.SetValue("")
			_ = m.auditRecord(argv)
			return m, executeCmd(m.executor, argv, m.generation)
		case 'n', 'N':
			m.pendingCommand = nil
			m.setStatus("Command cancelled")
			return m, nil
		}
	}
	if msg.Type == tea.KeyEsc {
		m.pendingCommand = nil
		m.setStatus("Command cancelled")
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft || m.showHelpModal {
		return m, nil
	}

	navStartY := 1
	if msg.Y >= navStartY && msg.Y < navStartY+len(navItems) {
		m.setView(navItems[msg.Y-navStartY])
		return m, nil
	}

	if m.activeView == ViewConnections && len(m.cfg.Profiles) > 0 {
		profileStartY := 5
		if msg.Y >= profileStartY && msg.Y < profileStartY+len(m.cfg.Profiles) {
			m.selectedProfile = msg.Y - profileStartY
			m.setStatus("Selected " + m.cfg.Profiles[m.selectedProfile].Name)
		}
	}

	if m.activeView == ViewKeys && len(m.keys) > 0 {
		keyListStartY := 5
		if msg.Y >= keyListStartY && msg.Y < keyListStartY+len(m.keys) {
			m.selectedKey = msg.Y - keyListStartY
			m.setStatus("Selected " + m.keys[m.selectedKey].Name)
		}
	}
	return m, nil
}

func (m Model) executeLine(line string) (tea.Model, tea.Cmd) {
	argv, err := command.Parse(line)
	if err != nil {
		m.setStatus("Parse error: " + err.Error())
		return m, nil
	}
	if len(argv) == 0 {
		return m, nil
	}
	if err := m.policy.Check(argv); err != nil {
		m.setStatus(err.Error())
		return m, nil
	}
	m.history.Add(strings.Join(security.RedactCommand(argv), " "))
	m.commandInput.SetValue("")
	if m.policy.RequiresConfirmation(argv) {
		m.pendingCommand = argv
		m.setStatus("Critical command: press y to confirm, n to cancel")
		return m, nil
	}

	_ = m.auditRecord(argv)
	m.setStatus("Executing " + strings.ToUpper(argv[0]))
	return m, executeCmd(m.executor, argv, m.generation)
}

func (m Model) handleConnected(msg connectedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		if msg.executor != nil {
			_ = msg.executor.Close()
		}
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Connect failed: " + msg.err.Error())
		return m, nil
	}
	m.closeClient()
	m.executor = msg.executor
	m.cfg.ActiveProfile = msg.profile.Name
	m.adapter = adapters.New(m.executor)
	m.opsManager = ops.New(m.executor)
	m.streamHub = stream.NewHub(msg.executor.Client())
	m.keys = nil
	m.selectedKey = 0
	m.selectedValue = result.Value{}
	m.setStatus("Connected to " + msg.profile.Name + " (" + msg.profile.Addr + ")")
	m.setView(ViewKeys)
	return m, m.refreshKeysCmd()
}

func (m Model) handleCommandExecuted(msg commandExecutedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Command failed: " + msg.err.Error())
		return m, nil
	}
	m.lastResult = msg.value
	m.setStatus("OK: " + msg.line)
	return m, nil
}

func (m *Model) nextView(delta int) {
	count := len(navItems)
	current := 0
	for i, view := range navItems {
		if view == m.activeView {
			current = i
			break
		}
	}
	m.setView(navItems[(current+delta+count)%count])
}

func (m *Model) selectPreviousProfile() {
	if m.selectedProfile > 0 {
		m.selectedProfile--
	}
}

func (m *Model) selectNextProfile() {
	if m.selectedProfile < len(m.cfg.Profiles)-1 {
		m.selectedProfile++
	}
}

func (m *Model) closeClient() {
	m.generation++
	if m.subscription != nil {
		_ = m.subscription.Close()
		m.subscription = nil
	}
	if m.monitor != nil {
		_ = m.monitor.Close()
		m.monitor = nil
	}
	if m.executor != nil {
		_ = m.executor.Close()
		m.executor = nil
	}

	m.adapter = nil
	m.opsManager = nil
	m.streamHub = nil
	m.keys = nil
	m.keyCursor = 0
	m.keyHasMore = false
	m.selectedKey = 0
	m.selectedValue = result.Value{}
	m.lastResult = result.Value{}
	m.pendingCommand = nil
	m.pendingDelete = nil
	m.filterMode = false
	m.opsSnapshot = ops.Snapshot{}
	m.streamMessages = nil
	m.monitorEvents = nil
}

func (m Model) handleKeysLoaded(msg keysLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Load keys failed: " + msg.err.Error())
		return m, nil
	}
	m.keys = msg.infos
	m.keyCursor = msg.cursor
	m.keyHasMore = msg.cursor != 0
	m.selectedKey = 0
	m.setStatus(fmt.Sprintf("Loaded %d keys", len(m.keys)))
	return m, nil
}

func (m Model) handleValueLoaded(msg valueLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Load value failed: " + msg.err.Error())
		return m, nil
	}
	m.selectedValue = msg.value
	m.setStatus("Loaded value for " + msg.key)
	return m, nil
}

func (m Model) handleKeysDeleted(msg keysDeletedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Delete failed: " + msg.err.Error())
		return m, nil
	}
	m.setStatus(fmt.Sprintf("Deleted %d key(s)", msg.deleted))
	return m, m.refreshKeysCmd()
}

func (m Model) handleKeysViewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.filterMode {
		if msg.Type == tea.KeyEnter {
			m.filterMode = false
			m.filterInput.Blur()
			return m, m.refreshKeysCmd()
		}
		if msg.Type == tea.KeyEsc {
			m.filterMode = false
			m.filterInput.Blur()
			m.filterInput.SetValue("")
			return m, m.refreshKeysCmd()
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		return m, cmd
	}

	switch msg.Type {
	case tea.KeyUp:
		m.selectPreviousKey()
		return m, nil
	case tea.KeyDown:
		m.selectNextKey()
		return m, nil
	case tea.KeyEnter:
		if m.selectedKey >= 0 && m.selectedKey < len(m.keys) {
			return m, loadValueCmd(m.adapter, m.keys[m.selectedKey], m.generation)
		}
		return m, nil
	}

	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return m, nil
	}
	switch msg.Runes[0] {
	case 'k':
		m.selectPreviousKey()
	case 'j':
		m.selectNextKey()
	case 'r':
		return m, m.refreshKeysCmd()
	case 'n':
		if m.keyHasMore {
			return m, m.keysPageCmd(m.keyCursor)
		}
	case '/':
		m.filterMode = true
		return m, m.filterInput.Focus()
	case 'd':
		if m.readOnly {
			m.setStatus("Delete is blocked in read-only mode")
			return m, nil
		}
		if m.selectedKey >= 0 && m.selectedKey < len(m.keys) {
			m.pendingDelete = []string{m.keys[m.selectedKey].Name}
			m.setStatus("Confirm delete: press y or n")
		}
	}
	return m, nil
}

func (m Model) handleDeleteConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case 'y', 'Y':
			keys := m.pendingDelete
			m.pendingDelete = nil
			_ = m.auditRecord(append([]string{"DEL"}, keys...))
			return m, deleteKeysCmd(m.adapter, keys, m.generation)
		case 'n', 'N':
			m.pendingDelete = nil
			m.setStatus("Delete cancelled")
			return m, nil
		}
	}
	if msg.Type == tea.KeyEsc {
		m.pendingDelete = nil
		m.setStatus("Delete cancelled")
	}
	return m, nil
}

func (m Model) refreshKeysCmd() tea.Cmd {
	return m.keysPageCmd(0)
}

func (m Model) keysPageCmd(cursor uint64) tea.Cmd {
	if m.adapter == nil {
		return nil
	}
	match := m.filterInput.Value()
	if match == "" {
		match = "*"
	}
	count := int64(m.cfg.PageSize)
	if count <= 0 {
		count = 100
	}
	return loadKeysCmd(m.adapter, match, count, cursor, m.generation)
}

func (m *Model) selectPreviousKey() {
	if m.selectedKey > 0 {
		m.selectedKey--
	}
}

func (m *Model) selectNextKey() {
	if m.selectedKey < len(m.keys)-1 {
		m.selectedKey++
	}
}

func (m Model) handleOpsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'r' {
		return m, loadOpsCmd(m.opsManager, m.generation)
	}
	return m, nil
}

func (m Model) handleOpsLoaded(msg opsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Load snapshot failed: " + msg.err.Error())
		return m, nil
	}
	m.opsSnapshot = msg.snapshot
	m.setStatus("Loaded ops snapshot")
	return m, nil
}

func (m Model) handleStreamsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return m, nil
	}

	switch msg.Runes[0] {
	case 's':
		if err := m.policy.Check([]string{"SUBSCRIBE"}); err != nil {
			m.setStatus(err.Error())
			return m, nil
		}
		if m.streamHub == nil {
			m.setStatus("Connect to Redis before subscribing")
			return m, nil
		}
		channel := m.streamChannel()
		m.setStatus("Subscribing to " + channel)
		return m, subscribeCmd(m.streamHub, channel, m.generation)
	case 'p':
		if err := m.policy.Check([]string{"PUBLISH"}); err != nil {
			m.setStatus(err.Error())
			return m, nil
		}
		if m.streamHub == nil {
			m.setStatus("Connect to Redis before publishing")
			return m, nil
		}
		return m, publishCmd(m.streamHub, m.streamChannel(), "hello from redis-cockpit", m.generation)
	case 'm':
		if err := m.policy.Check([]string{"MONITOR"}); err != nil {
			m.setStatus(err.Error())
			return m, nil
		}
		if m.streamHub == nil {
			m.setStatus("Connect to Redis before monitoring")
			return m, nil
		}
		m.setStatus("Starting MONITOR stream")
		return m, startMonitorCmd(m.streamHub, m.generation)
	case 'x':
		m.generation++
		if m.subscription != nil {
			_ = m.subscription.Close()
			m.subscription = nil
		}
		if m.monitor != nil {
			_ = m.monitor.Close()
			m.monitor = nil
		}
		m.setStatus("Streams stopped")
		return m, nil
	}

	var cmd tea.Cmd
	m.streamInput, cmd = m.streamInput.Update(msg)
	return m, cmd
}

func (m Model) streamChannel() string {
	channel := strings.TrimSpace(m.streamInput.Value())
	if channel == "" {
		return "redis-cockpit"
	}
	return channel
}

func (m Model) handleSubscribed(msg subscribedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		if msg.subscription != nil {
			_ = msg.subscription.Close()
		}
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Subscribe failed: " + msg.err.Error())
		return m, nil
	}
	if m.subscription != nil {
		_ = m.subscription.Close()
	}
	m.subscription = msg.subscription
	m.streamMessages = nil
	m.setStatus("Subscribed to " + m.streamChannel())
	return m, waitStreamMessageCmd(m.subscription, m.generation)
}

func (m Model) handlePublished(msg publishedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("Publish failed: " + msg.err.Error())
		return m, nil
	}
	m.setStatus("Published to " + m.streamChannel())
	return m, nil
}

func (m Model) handleMonitorStarted(msg monitorStartedMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		if msg.monitor != nil {
			_ = msg.monitor.Close()
		}
		return m, nil
	}
	if msg.err != nil {
		m.setStatus("MONITOR failed: " + msg.err.Error())
		return m, nil
	}
	if m.monitor != nil {
		_ = m.monitor.Close()
	}
	m.monitor = msg.monitor
	m.monitorEvents = nil
	m.setStatus("MONITOR started")
	return m, waitMonitorEventCmd(m.monitor, m.generation)
}

func (m Model) handleStreamMessage(msg streamMessageMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	m.streamMessages = append(m.streamMessages, msg.message)
	if len(m.streamMessages) > 100 {
		m.streamMessages = m.streamMessages[len(m.streamMessages)-100:]
	}
	m.setStatus("Pub/Sub: " + msg.message.Channel)
	return m, waitStreamMessageCmd(m.subscription, m.generation)
}

func (m Model) auditRecord(argv []string) error {
	if m.audit == nil {
		return nil
	}
	profileName := m.cfg.ActiveProfile
	if profile, ok := m.cfg.FindProfile(profileName); ok {
		profileName = profile.Name
	}
	return m.audit.Record(profileName, argv)
}

func (m Model) handleMonitorEvent(msg monitorEventMsg) (tea.Model, tea.Cmd) {
	if msg.gen != m.generation {
		return m, nil
	}
	m.monitorEvents = append(m.monitorEvents, msg.event)
	if len(m.monitorEvents) > 100 {
		m.monitorEvents = m.monitorEvents[len(m.monitorEvents)-100:]
	}
	m.setStatus("MONITOR event received")
	return m, waitMonitorEventCmd(m.monitor, m.generation)
}
