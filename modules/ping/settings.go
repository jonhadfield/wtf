package ping

import (
	"github.com/olebedev/config"
	"github.com/wtfutil/wtf/cfg"
	"github.com/wtfutil/wtf/utils"
)

const (
	defaultFocusable = true
	defaultTitle     = "Ping"
)

type Settings struct {
	common      *cfg.Common
	targets     []string
	pingTimeout int
	showIP      bool
	useEmoji    bool
	format      bool
	formatStyle string
	wrapText    bool
}

func NewSettingsFromYAML(name string, ymlConfig *config.Config, globalConfig *config.Config) *Settings {
	settings := Settings{
		common:      cfg.NewCommonSettingsFromModule(name, defaultTitle, defaultFocusable, ymlConfig, globalConfig),
		targets:     utils.ToStrs(ymlConfig.UList("targets")),
		showIP:      ymlConfig.UBool("showIP", true),
		pingTimeout: ymlConfig.UInt("pingTimeout", 4),
		useEmoji:    ymlConfig.UBool("useEmoji", true),
	}

	return &settings
}
