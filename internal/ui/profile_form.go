package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/Ryequ/redis-cockpit/internal/redis"
)

type profileIOMode int

const (
	profileIOImport profileIOMode = iota
	profileIOExport
)

type profileIOForm struct {
	mode      profileIOMode
	pathInput textinput.Model
}

func newProfileIOForm(mode profileIOMode) profileIOForm {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = "/path/to/profiles.yaml"
	input.CharLimit = 1024
	input.Focus()
	return profileIOForm{mode: mode, pathInput: input}
}

func (f profileIOForm) View() string {
	title := "Import Profiles"
	if f.mode == profileIOExport {
		title = "Export Profiles"
	}
	return strings.Join([]string{
		title,
		"",
		"Path: " + f.pathInput.View(),
		"",
		"[OK]  [Cancel]",
		"Enter 确认  •  Esc 取消",
	}, "\n")
}

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
	profileFieldSSHHost
	profileFieldSSHPort
	profileFieldSSHUser
	profileFieldSSHKeyPath
	profileFieldCloudProvider
	profileFieldAuthMechanism
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
	sshPort := profile.SSHPort
	if sshPort <= 0 {
		sshPort = 22
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
		profile.SSHHost,
		strconv.Itoa(sshPort),
		profile.SSHUser,
		profile.SSHKeyPath,
		profile.CloudProvider,
		profile.AuthMechanism,
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
		"optional SSH bastion host",
		"22",
		"SSH user",
		"optional private key path",
		"aws/gcp/azure/aliyun/tencent",
		"default/iam",
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
			"SSH Host",
			"SSH Port",
			"SSH User",
			"SSH Key Path",
			"Cloud Provider",
			"Auth Mechanism",
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
	sshPort, err := parseIntDefault(f.inputs[profileFieldSSHPort].Value(), 22)
	if err != nil {
		return redis.ConnectionProfile{}, fmt.Errorf("ssh port must be an integer")
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
		SSHHost:            strings.TrimSpace(f.inputs[profileFieldSSHHost].Value()),
		SSHPort:            sshPort,
		SSHUser:            strings.TrimSpace(f.inputs[profileFieldSSHUser].Value()),
		SSHKeyPath:         strings.TrimSpace(f.inputs[profileFieldSSHKeyPath].Value()),
		CloudProvider:      strings.ToLower(strings.TrimSpace(f.inputs[profileFieldCloudProvider].Value())),
		AuthMechanism:      strings.ToLower(strings.TrimSpace(f.inputs[profileFieldAuthMechanism].Value())),
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
	lines := []string{
		title,
		"",
		"Templates: [AWS] [GCP] [Azure] [Aliyun] [Tencent]",
		"[Save]  [Cancel]",
		"",
	}
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

func (f *profileForm) applyCloudTemplate(provider string) {
	preset := redis.ApplyCloudPreset(redis.ConnectionProfile{}, provider)
	f.inputs[profileFieldCloudProvider].SetValue(preset.CloudProvider)
	f.inputs[profileFieldAuthMechanism].SetValue(preset.AuthMechanism)
	f.bools[profileBoolTLS] = preset.TLS
}

func parseIntDefault(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}
