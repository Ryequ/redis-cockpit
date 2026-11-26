package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	goredis "github.com/redis/go-redis/v9"

	"github.com/Ryequ/redis-cockpit/internal/command"
	"github.com/Ryequ/redis-cockpit/internal/config"
	"github.com/Ryequ/redis-cockpit/internal/ops"
	"github.com/Ryequ/redis-cockpit/internal/redis"
	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
	"github.com/Ryequ/redis-cockpit/internal/result"
	"github.com/Ryequ/redis-cockpit/internal/security"
	"github.com/Ryequ/redis-cockpit/internal/stream"
)

// ViewID identifies a top-level workspace view.
type ViewID int

const (
	ViewConnections ViewID = iota
	ViewKeys
	ViewWorkbench
	ViewOps
	ViewStreams
)

func (v ViewID) String() string {
	switch v {
	case ViewConnections:
		return "Connections"
	case ViewKeys:
		return "Keys"
	case ViewWorkbench:
		return "Workbench"
	case ViewOps:
		return "Ops"
	case ViewStreams:
		return "Streams"
	default:
		return "Unknown"
	}
}

// Model is the root Bubble Tea model. It owns global chrome and view routing.
type Model struct {
	width         int
	height        int
	activeView    ViewID
	mouse         bool
	status        string
	showHelpModal bool

	cfg                  config.Config
	configPath           string
	profileForm          *profileForm
	profileIOForm        *profileIOForm
	pendingProfileDelete string
	readOnly             bool
	policy               security.Policy
	selectedProfile      int
	profileSearch        textinput.Model
	profileSearchMode    bool
	profileSort          profileSortKey
	profileGroupByMode   bool
	profileHealth        map[string]ProfileHealth
	profileGeneration    uint64
	batchTestTotal       int
	batchTestDone        int
	batchTestOK          int
	batchTestFail        int
	client               goredis.UniversalClient
	connectionClose      func() error
	executor             *redis.Executor
	adapter              *adapters.Manager
	history              *command.History
	commandInput         textinput.Model
	lastResult           result.Value
	pendingCommand       []string

	keys            []adapters.KeyInfo
	keyCursor       uint64
	selectedKey     int
	selectedValue   result.Value
	filterInput     textinput.Model
	filterMode      bool
	pendingDelete   []string
	keyForm         *keyForm
	pendingKeyFocus string
	keyTreeMode     bool
	keyTreeExpanded map[string]bool
	keyScroll       int

	opsManager     *ops.Manager
	opsSnapshot    ops.Snapshot
	streamHub      *stream.Hub
	streamInput    textinput.Model
	subscription   *stream.Subscription
	monitor        *stream.Monitor
	streamMessages []stream.Message
	monitorEvents  []stream.MonitorEvent

	audit      *security.AuditLog
	keyHasMore bool
	generation uint64
}

func NewModel() Model {
	return NewModelWithConfig(config.Default(), false)
}

func NewModelWithConfig(cfg config.Config, readOnly bool) Model {
	input := textinput.New()
	input.Placeholder = "SET greeting hello"
	input.Prompt = "> "
	input.CharLimit = 4096

	filter := textinput.New()
	filter.Placeholder = "*"
	filter.Prompt = "/"
	filter.CharLimit = 256

	profileSearch := textinput.New()
	profileSearch.Placeholder = "filter profiles"
	profileSearch.Prompt = "/"
	profileSearch.CharLimit = 256

	streamChannel := textinput.New()
	streamChannel.Placeholder = "redis-cockpit"
	streamChannel.Prompt = "#"
	streamChannel.CharLimit = 256

	selected := 0
	for i, profile := range cfg.Profiles {
		if profile.Name == cfg.ActiveProfile {
			selected = i
			break
		}
	}

	return Model{
		width:           100,
		height:          30,
		activeView:      ViewConnections,
		mouse:           true,
		status:          "Ready",
		cfg:             cfg,
		readOnly:        readOnly,
		policy:          security.NewPolicy(readOnly),
		selectedProfile: selected,
		profileSearch:   profileSearch,
		profileHealth:   make(map[string]ProfileHealth),
		keyTreeExpanded: make(map[string]bool),
		history:         command.NewHistory(100),
		commandInput:    input,
		filterInput:     filter,
		streamInput:     streamChannel,
	}
}

func (m Model) Width() int         { return m.width }
func (m Model) Height() int        { return m.height }
func (m Model) ActiveView() ViewID { return m.activeView }
func (m Model) MouseEnabled() bool { return m.mouse }

func (m *Model) SetMouseEnabled(enabled bool) {
	m.mouse = enabled
}

func (m *Model) SetAuditLog(audit *security.AuditLog) {
	m.audit = audit
}

func (m *Model) SetConfigPath(path string) {
	m.configPath = path
}

func (m Model) Status() string { return m.status }

func (m *Model) setStatus(status string) {
	m.status = status
}

func (m *Model) setView(view ViewID) {
	if view < ViewConnections || view > ViewStreams {
		return
	}
	m.activeView = view
	m.setStatus(view.String() + " view")
}
