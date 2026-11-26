package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

type profileFormMode int

const (
	profileCreate profileFormMode = iota
	profileEdit
)

const (
	profileFieldName int = iota
	profileFieldAddr
	profileFieldUsername
	profileFieldPassword
	profileFieldDB
	profileFieldSentinelAddrs
	profileFieldMasterName
	profileFieldTimeout
	profileFieldCount
)

const (
	profileBoolTLS int = iota
	profileBoolInsecure
	profileBoolCluster
	profileBoolSentinel
	profileBoolReadOnly
	profileBoolCount
)

type profileForm struct {
	mode         profileFormMode
	originalName string
	originalID   string
	inputs       []textinput.Model
	labels       []string
	bools        []bool
	boolLabels   []string
	focus        int
}

func newProfileForm(profile redis.ConnectionProfile, mode profileFormMode) profileForm {
	timeout := profile.TimeoutSeconds
	if timeout <= 0 {
		timeout = 5
	}

	values := []string{
		profile.Name,
		profile.Addr,
		profile.Username,
		profile.Password,
		strconv.Itoa(profile.DB),
		strings.Join(profile.SentinelAddrs, ","),
		profile.MasterName,
		strconv.Itoa(timeout),
	}
	placeholders := []string{
		"profile name",
		"host:port",
		"optional",
		"optional or ${ENV_VAR}",
		"0",
		"host:port,host:port",
		"sentinel master name",
		"5",
	}

	form := profileForm{
		mode:         mode,
		originalName: profile.Name,
		originalID:   profile.ID,
		labels: []string{
			"Name",
			"Address",
			"Username",
			"Password",
			"DB",
			"Sentinel Addresses",
			"Master Name",
			"Timeout (s)",
		},
		bools: []bool{
			profile.TLS,
			profile.InsecureSkipVerify,
			profile.Cluster,
			profile.Sentinel,
			profile.ReadOnly,
		},
		boolLabels: []string{
			"TLS",
			"Skip TLS Verify",
			"Cluster",
			"Sentinel",
			"Read Only",
		},
	}
	for i, value := range values {
		input := textinput.New()
		input.SetValue(value)
		input.Placeholder = placeholders[i]
		input.CharLimit = 512
		if i == profileFieldPassword {
			input.EchoMode = textinput.EchoPassword
		}
		form.inputs = append(form.inputs, input)
	}
	form.inputs[0].Focus()
	return form
}

func (f *profileForm) focusCount() int {
	return profileFieldCount + profileBoolCount
}

func (f *profileForm) setFocus(next int) {
	if f.focus < profileFieldCount {
		f.inputs[f.focus].Blur()
	}
	f.focus = next
	if f.focus < profileFieldCount {
		f.inputs[f.focus].Focus()
	}
}

func (f *profileForm) nextFocus() {
	f.setFocus((f.focus + 1) % f.focusCount())
}

func (f *profileForm) prevFocus() {
	f.setFocus((f.focus - 1 + f.focusCount()) % f.focusCount())
}

func (f *profileForm) toggleBool() {
	if f.focus >= profileFieldCount && f.focus < f.focusCount() {
		idx := f.focus - profileFieldCount
		f.bools[idx] = !f.bools[idx]
	}
}

func (f profileForm) profile() (redis.ConnectionProfile, error) {
	db, err := parseIntDefault(f.inputs[profileFieldDB].Value(), 0)
	if err != nil {
		return redis.ConnectionProfile{}, fmt.Errorf("db must be an integer")
	}
	timeout, err := parseIntDefault(f.inputs[profileFieldTimeout].Value(), 5)
	if err != nil {
		return redis.ConnectionProfile{}, fmt.Errorf("timeout must be an integer")
	}
	if timeout < 0 {
		return redis.ConnectionProfile{}, fmt.Errorf("timeout must be non-negative")
	}

	profile := redis.ConnectionProfile{
		ID:                 f.originalID,
		Name:               strings.TrimSpace(f.inputs[profileFieldName].Value()),
		Addr:               strings.TrimSpace(f.inputs[profileFieldAddr].Value()),
		Username:           strings.TrimSpace(f.inputs[profileFieldUsername].Value()),
		Password:           f.inputs[profileFieldPassword].Value(),
		DB:                 db,
		TLS:                f.bools[profileBoolTLS],
		InsecureSkipVerify: f.bools[profileBoolInsecure],
		Cluster:            f.bools[profileBoolCluster],
		Sentinel:           f.bools[profileBoolSentinel],
		ReadOnly:           f.bools[profileBoolReadOnly],
		TimeoutSeconds:     timeout,
	}
	for _, addr := range strings.Split(f.inputs[profileFieldSentinelAddrs].Value(), ",") {
		if addr = strings.TrimSpace(addr); addr != "" {
			profile.SentinelAddrs = append(profile.SentinelAddrs, addr)
		}
	}
	profile.MasterName = strings.TrimSpace(f.inputs[profileFieldMasterName].Value())
	if profile.ID == "" {
		profile.ID = profile.Name
	}

	if err := profile.Validate(); err != nil {
		return redis.ConnectionProfile{}, err
	}
	return profile, nil
}

func (f profileForm) View() string {
	title := "New Connection Profile"
	if f.mode == profileEdit {
		title = "Edit Connection Profile"
	}
	lines := []string{title, ""}
	for i, label := range f.labels {
		marker := "  "
		if i == f.focus {
			marker = "▸ "
		}
		lines = append(lines, fmt.Sprintf("%s%-20s %s", marker, label+":", f.inputs[i].View()))
	}
	for i, label := range f.boolLabels {
		focus := f.focus - profileFieldCount
		marker := "  "
		if i == focus {
			marker = "▸ "
		}
		value := "false"
		if f.bools[i] {
			value = "true"
		}
		lines = append(lines, fmt.Sprintf("%s%-20s %s", marker, label+":", value))
	}
	lines = append(lines, "", "Tab/Shift+Tab 切换字段  •  Space 切换开关  •  Enter 保存  •  Esc 取消")
	return strings.Join(lines, "\n")
}

func parseIntDefault(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}
