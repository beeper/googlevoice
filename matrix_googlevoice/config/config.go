package config

import "maunium.net/go/mautrix/bridge/bridgeconfig"

type Config struct {
	*bridgeconfig.BaseConfig `yaml:",inline"`

	Bridge BridgeConfig `yaml:"bridge"`
}
