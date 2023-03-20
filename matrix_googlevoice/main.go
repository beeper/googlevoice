package main

import (
	_ "embed"
	"github.com/emostar/libgooglevoice/matrix_googlevoice/config"
	"maunium.net/go/mautrix/bridge"
	"maunium.net/go/mautrix/bridge/bridgeconfig"
	"maunium.net/go/mautrix/bridge/commands"
	"maunium.net/go/mautrix/bridge/status"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/util/configupgrade"
	"sync"
)

//go:embed config.yaml
var ExampleConfig string

type GVBridge struct {
	bridge.Bridge

	Config *config.Config

	usersByMXID   map[id.UserID]*User
	usersLock     sync.Mutex
	portalsByMXID map[id.RoomID]*Portal
	portalsLock   sync.Mutex
}

func (br *GVBridge) GetExampleConfig() string {
	return ExampleConfig
}

func (br *GVBridge) GetConfigPtr() interface{} {
	br.Config = &config.Config{
		BaseConfig: &br.Bridge.Config,
	}
	br.Config.BaseConfig.Bridge = &br.Config.Bridge
	return br.Config
}

func (br *GVBridge) Start() {
	br.Log.Info("Starting Google Voice bridge")
	br.Init()
}

func (br *GVBridge) Stop() {
	//TODO implement me
	panic("implement Stop")
}

func (br *GVBridge) IsGhost(id id.UserID) bool {
	//TODO implement me
	br.Log.Infof("Got IsGhost request for %#v (%s)", id, id)
	return false
}

func (br *GVBridge) GetIGhost(id id.UserID) bridge.Ghost {
	//TODO implement me
	panic("implement GetIGhost")
}

func (br *GVBridge) CreatePrivatePortal(id id.RoomID, user bridge.User, ghost bridge.Ghost) {
	//TODO implement me
	panic("implement CreatePrivatePortal")
}

func (br *GVBridge) Init() {
	br.CommandProcessor = commands.NewProcessor(&br.Bridge)
	br.RegisterCommands()

	br.SendGlobalBridgeState(status.BridgeState{StateEvent: status.StateUnconfigured}.Fill(nil))
}

func main() {
	br := &GVBridge{
		usersByMXID:   make(map[id.UserID]*User),
		portalsByMXID: make(map[id.RoomID]*Portal),
	}

	br.Bridge = bridge.Bridge{
		Name:         "emostar-googlevoice",
		URL:          "https://github.com/emostar/matrix-googlevoice",
		Description:  "A Matrix-Google Voice puppeting bridge",
		Version:      "0.0.1",
		ProtocolName: "GoogleVoice",

		ConfigUpgrader: &configupgrade.StructUpgrader{
			SimpleUpgrader: configupgrade.SimpleUpgrader(func(helper *configupgrade.Helper) {
				bridgeconfig.Upgrader.DoUpgrade(helper)
			}),
			Blocks: [][]string{},
			Base:   ExampleConfig,
		},

		Child: br,
	}

	br.Main()
}
