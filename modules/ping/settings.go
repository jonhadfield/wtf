package ping

import (
	"github.com/olebedev/config"
	"github.com/wtfutil/wtf/cfg"
	"github.com/wtfutil/wtf/utils"
)

const (
	defaultFocusable = true
	defaultTitle     = "Ping"
	moduleName       = "Ping"
)

type Settings struct {
	common      *cfg.Common
	targets     []string
	pingTimeout int
	showIP      bool
	privileged  bool
	useEmoji    bool
	logging     bool
	wrapText    bool
}

func NewSettingsFromYAML(name string, ymlConfig *config.Config, globalConfig *config.Config) *Settings {
	settings := Settings{
		common:      cfg.NewCommonSettingsFromModule(name, defaultTitle, defaultFocusable, ymlConfig, globalConfig),
		targets:     utils.ToStrs(ymlConfig.UList("targets")),
		showIP:      ymlConfig.UBool("showIP", true),
		pingTimeout: ymlConfig.UInt("pingTimeout", 200),
		useEmoji:    ymlConfig.UBool("useEmoji", true),
		logging:     ymlConfig.UBool("logging", false),
		privileged:  ymlConfig.UBool("privileged", false),
	}

	return &settings
}
