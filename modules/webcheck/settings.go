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
	common                *cfg.Common
	urls                  []string
	warnCodes             []int
	tlsHandshakeTimeout   int
	responseHeaderTimeout int
	fullResponseTimeout   int
	followRedirects       bool
	ignoreBadSSL          bool
	useEmoji              bool
	format                bool
	formatStyle           string
	wrapText              bool
}

func NewSettingsFromYAML(name string, ymlConfig *config.Config, globalConfig *config.Config) *Settings {
	settings := Settings{
		common:                cfg.NewCommonSettingsFromModule(name, defaultTitle, defaultFocusable, ymlConfig, globalConfig),
		urls:                  utils.ToStrs(ymlConfig.UList("urls")),
		warnCodes:             utils.ToInts(ymlConfig.UList("warnCodes")),
		tlsHandshakeTimeout:   ymlConfig.UInt("tlsHandshakeTimout", 2),
		responseHeaderTimeout: ymlConfig.UInt("responseHeaderTimeout", 3),
		fullResponseTimeout:   ymlConfig.UInt("fullResponseTimeout", 6),
		followRedirects:       ymlConfig.UBool("followRedirects", true),
		ignoreBadSSL:          ymlConfig.UBool("ignoreBadSSL", false),
		useEmoji:              ymlConfig.UBool("useEmoji", true),
	}

	return &settings
}
