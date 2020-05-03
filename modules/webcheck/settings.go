package webcheck

import (
	"github.com/olebedev/config"
	"github.com/wtfutil/wtf/cfg"
	"github.com/wtfutil/wtf/utils"
)

const (
	defaultFocusable = true
	defaultTitle     = "WebCheck"
	moduleName       = "WebCheck"
)

type Settings struct {
	common          *cfg.Common
	urls            []string
	warnCodes       []int
	followRedirects bool
	ignoreBadSSL    bool
	useEmoji        bool
	format          bool
	formatStyle     string
	wrapText        bool
}

func NewSettingsFromYAML(name string, ymlConfig *config.Config, globalConfig *config.Config) *Settings {
	settings := Settings{
		common:          cfg.NewCommonSettingsFromModule(name, defaultTitle, defaultFocusable, ymlConfig, globalConfig),
		urls:            utils.ToStrs(ymlConfig.UList("urls")),
		warnCodes:       utils.ToInts(ymlConfig.UList("warnCodes")),
		followRedirects: ymlConfig.UBool("followRedirects", true),
		ignoreBadSSL:    ymlConfig.UBool("ignoreBadSSL", false),
		useEmoji:        ymlConfig.UBool("useEmoji", true),
	}

	return &settings
}
