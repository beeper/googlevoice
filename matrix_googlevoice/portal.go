package main

import (
	"fmt"
	log "maunium.net/go/maulogger/v2"
	"maunium.net/go/mautrix/appservice"
	"maunium.net/go/mautrix/bridge"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type PortalMessage struct {
}

type PortalMatrixMessage struct {
}

type Portal struct {
	bridge *GVBridge
	log    log.Logger

	messages       chan PortalMessage
	matrixMessages chan PortalMatrixMessage
}

func (p *Portal) IsPrivateChat() bool {
	p.log.Info("Got IsPrivateChat request")
	//TODO implement me
	return false
}

func (p *Portal) MarkEncrypted() {
	p.log.Info("Got MarkEncrypted request")
	//TODO implement me
}

func (p *Portal) MainIntent() *appservice.IntentAPI {
	p.log.Info("Got MainIntent request")
	//TODO implement me
	return nil
}

func (p *Portal) ReceiveMatrixEvent(user bridge.User, evt *event.Event) {
	p.log.Infof("Got ReceiveMatrixEvent request %s %#v", user, evt)
	//TODO implement me with the matrixMessages channel
	//p.matrixMessages <- PortalMatrixMessage{}
}

func (p *Portal) UpdateBridgeInfo() {
	p.log.Info("Got UpdateBridgeInfo request")
	//TODO implement me
}

func (p *Portal) IsEncrypted() bool {
	p.log.Info("Got IsEncrypted request")
	return false
}

func (br *GVBridge) newPortal(mxid id.RoomID) *Portal {
	return &Portal{
		bridge:         br,
		log:            br.Log.Sub(fmt.Sprintf("Portal/%s", mxid)),
		messages:       make(chan PortalMessage, 10),
		matrixMessages: make(chan PortalMatrixMessage, 10),
	}
}

func (br *GVBridge) GetPortal(mxid id.RoomID) *Portal {
	br.portalsLock.Lock()
	defer br.portalsLock.Unlock()

	portal, ok := br.portalsByMXID[mxid]
	if !ok {
		portal = br.newPortal(mxid)
		br.portalsByMXID[mxid] = portal
	}
	return portal
}

func (br *GVBridge) GetIPortal(mxid id.RoomID) bridge.Portal {
	br.Log.Info("Got GetIPortal request", mxid)
	return br.GetPortal(mxid)
}

func (br *GVBridge) GetAllIPortals() []bridge.Portal {
	br.Log.Info("Got GetAllIPortals request")
	//TODO implement me
	panic("implement GetAllIPortals")
}
