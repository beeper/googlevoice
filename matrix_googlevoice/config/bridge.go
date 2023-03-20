package config

import (
	"maunium.net/go/mautrix/bridge/bridgeconfig"
)

type BridgeConfig struct {
	UsernameTemplate    string `yaml:"username_template"`
	DisplaynameTemplate string `yaml:"displayname_template"`

	MessageStatusEvents bool `yaml:"message_status_events"`

	CommandPrefix      string                           `yaml:"command_prefix"`
	ManagementRoomText bridgeconfig.ManagementRoomTexts `yaml:"management_room_text"`

	Encryption bridgeconfig.EncryptionConfig `yaml:"encryption"`

	Permissions bridgeconfig.PermissionConfig `yaml:"permissions"`
}

func (bc BridgeConfig) GetResendBridgeInfo() bool {
	//TODO implement me
	return false
}

func (bc BridgeConfig) EnableMessageStatusEvents() bool {
	//TODO implement me
	return false
}

func (bc BridgeConfig) EnableMessageErrorNotices() bool {
	//TODO implement me
	return false
}

func (bc BridgeConfig) Validate() error {
	//TODO implement me
	return nil
}

func (bc BridgeConfig) GetEncryptionConfig() bridgeconfig.EncryptionConfig {
	return bc.Encryption
}

func (bc BridgeConfig) GetCommandPrefix() string {
	return bc.CommandPrefix
}

func (bc BridgeConfig) GetManagementRoomTexts() bridgeconfig.ManagementRoomTexts {
	return bc.ManagementRoomText
}

func (bc BridgeConfig) FormatUsername(userID string) string {
	return ""
}
