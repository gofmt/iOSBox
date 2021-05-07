package idevice

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/xerrors"
)

type ForwardService struct {
	entry  *DeviceEntry
	listen net.Listener
	conn   *USBConn
}

func NewForwardService(entry *DeviceEntry) *ForwardService {
	return &ForwardService{entry: entry}
}

func (fs *ForwardService) Start(hostPort, remotePort uint16, cb func(string, error)) (err error) {
	fs.listen, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", hostPort))
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := fs.listen.Accept()
			if err != nil {
				cb(fmt.Sprintf("Error accepting new connection %v", err), nil)
				continue
			}
			cb(fmt.Sprintf("new client connected"), nil)
			go fs.startNewProxyConn(conn, fs.entry.DeviceID, remotePort, cb)
		}
	}()

	return nil
}

func (fs *ForwardService) Close() {
	_ = fs.listen.Close()
	if fs.conn != nil {
		fs.conn.Close()
	}
}

func (fs *ForwardService) startNewProxyConn(conn net.Conn, deviceId int, remotePort uint16, cb func(string, error)) {
	var err error
	fs.conn, err = NewUSBConn()
	if err != nil {
		cb("", xerrors.Errorf("could not connect to usbmuxd: %w", err))
		_ = conn.Close()
		return
	}

	if err := fs.conn.Connect(deviceId, Ntohs(remotePort)); err != nil {
		cb("", xerrors.Errorf("could not connect to remote: %w", err))
		_ = conn.Close()
		return
	}
	cb(fmt.Sprintf("Connected to port: %d", remotePort), nil)

	go func() {
		_, _ = io.Copy(conn, fs.conn.Conn.Reader())
	}()
	go func() {
		_, _ = io.Copy(fs.conn.Conn.Writer(), conn)
	}()
}
