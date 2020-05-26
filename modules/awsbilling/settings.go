package awsbilling

import (
	"github.com/olebedev/config"
	"github.com/wtfutil/wtf/cfg"
)

const (
	defaultFocusable = true
	defaultTitle     = "AWS Billing"
)

type Settings struct {
	common     *cfg.Common
	wrapText   bool
	accounts   []interface{}
	aliasWidth int
	output     string
}

func NewSettingsFromYAML(name string, ymlConfig *config.Config, globalConfig *config.Config) *Settings {
	settings := Settings{
		common:     cfg.NewCommonSettingsFromModule(name, defaultTitle, defaultFocusable, ymlConfig, globalConfig),
		accounts:   ymlConfig.UList("accounts"),
		aliasWidth: ymlConfig.UInt("aliasWidth", 10),
		output:     ymlConfig.UString("output", "default"),
	}

	return &settings
}
