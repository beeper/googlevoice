package main

import (
	"fmt"
	"github.com/emostar/libgooglevoice/libgooglevoice"
	"go.uber.org/zap"
	log "maunium.net/go/maulogger/v2"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/bridge"
	"maunium.net/go/mautrix/bridge/bridgeconfig"
	"maunium.net/go/mautrix/id"
)

type User struct {
	bridge *GVBridge
	log    log.Logger

	GVoice *libgooglevoice.GoogleVoiceClient

	PermissionLevel bridgeconfig.PermissionLevel

	// Database values
	MXID          id.UserID
	MangementRoom id.RoomID
	SpaceRoom     id.RoomID
}

func (u *User) GetPermissionLevel() bridgeconfig.PermissionLevel {
	u.log.Info("Got GetPermissionLevel request")
	return u.PermissionLevel
}

func (u *User) IsLoggedIn() bool {
	u.log.Info("Got IsLoggedIn request")
	//TODO implement me
	return true
}

func (u *User) GetManagementRoomID() id.RoomID {
	u.log.Info("Got GetManagementRoomID request")
	if len(u.MangementRoom) == 0 {
		resp, err := u.bridge.Bridge.Bot.CreateRoom(&mautrix.ReqCreateRoom{
			Topic:           "GoogleVoice bridge notices",
			IsDirect:        true,
			CreationContent: map[string]interface{}{},
		})
		if err != nil {
			u.log.Errorln("Failed to auto-create management room:", err)
		} else {
			u.SetManagementRoom(resp.RoomID)
		}
	}
	return u.MangementRoom
}

func (u *User) SetManagementRoom(roomID id.RoomID) {
	u.log.Info("Got SetManagementRoom request")
	//TODO implement me
	u.MangementRoom = roomID
}

func (u *User) GetMXID() id.UserID {
	u.log.Info("Got GetMXID request")
	return u.MXID
}

func (u *User) GetIDoublePuppet() bridge.DoublePuppet {
	u.log.Info("Got GetIDoublePuppet request")
	//TODO implement me
	panic("implement me")
}

func (u *User) GetIGhost() bridge.Ghost {
	u.log.Info("Got GetIGhost request")
	//TODO implement me
	panic("implement me")
}

func (br *GVBridge) GetUser(userID id.UserID, create bool) *User {
	br.usersLock.Lock()
	defer br.usersLock.Unlock()

	if user, ok := br.usersByMXID[userID]; ok {
		return user
	}

	if !create {
		return nil
	}

	zapLog, _ := zap.NewDevelopment()

	user := &User{
		bridge:          br,
		log:             br.Log.Sub(fmt.Sprintf("User/%s", userID)),
		PermissionLevel: bridgeconfig.PermissionLevelAdmin,
		GVoice:          libgooglevoice.NewGoogleVoiceClient(zapLog.Sugar()),
	}

	br.usersByMXID[userID] = user
	return user
}

func (br *GVBridge) GetIUser(userID id.UserID, create bool) bridge.User {
	return br.GetUser(userID, create)
}
