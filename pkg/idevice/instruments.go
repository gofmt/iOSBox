package idevice

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

type processControlDispatcher struct {
	conn *dtx.Connection
}

func (p processControlDispatcher) Dispatch(msg dtx.Message) {
	dtx.SendAckIfNeeded(p.conn, msg)
}

type InstrumentsService struct {
	channel *dtx.Channel
	conn    *dtx.Connection
}

func NewInstrumentsService(device ios.DeviceEntry) (*InstrumentsService, error) {
	dtxConn, err := dtx.NewConnection(device, "com.apple.instruments.remoteserver")
	if err != nil {
		dtxConn, err = dtx.NewConnection(device, "com.apple.instruments.remoteserver.DVTSecureSocketProxy")
		if err != nil {
			return nil, err
		}
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(
		"com.apple.instruments.server.services.deviceinfo",
		processControlDispatcher{dtxConn},
	)

	return &InstrumentsService{processControlChannel, dtxConn}, nil
}

func (d *InstrumentsService) Close() {
	d.conn.Close()
}
