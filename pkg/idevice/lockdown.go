package idevice

import (
	"golang.org/x/xerrors"
	"howett.net/plist"
)

type valutRequest struct {
	Label   string
	Key     string `plist:"Key,omitempty"`
	Request string
	Domain  string `plist:"Domain,omitempty"`
	Value   string `plist:"Value,omitempty"`
}

type ValueResponse struct {
	Key     string
	Request string
	Error   string
	Domain  string
	Value   interface{}
}

type NonVolatileRAMType struct {
	AutoBoot              []byte `plist:"auto-boot"`
	BacklightLevel        []byte `plist:"backlight-level"`
	BootArgs              string `plist:"boot-args"`
	Bootdelay             []byte `plist:"bootdelay"`
	ComAppleSystemTz0Size []byte `plist:"com.apple.System.tz0-size"`
	OblitBegins           []byte `plist:"oblit-begins"`
	Obliteration          []byte `plist:"obliteration"`
}

type BasebandKeyHashInformationType struct {
	AKeyStatus int
	SKeyHash   []byte
	SKeyStatus int
}

type LockdownConn struct {
	Conn      IConn
	sessionId string
}

func NewLockdownConn(conn IConn) *LockdownConn {
	return &LockdownConn{Conn: conn}
}

func (l *LockdownConn) Close() {
	l.Conn.Close()
}

func (l *LockdownConn) Send(msg interface{}) error {
	bs, err := l.Conn.Encode(msg)
	if err != nil {
		return err
	}

	return l.Conn.Write(bs)
}

func (l *LockdownConn) Recv() ([]byte, error) {
	return l.Conn.Decode(l.Conn.Reader())
}

type startSessionRequest struct {
	Label           string
	ProtocolVersion string
	Request         string
	HostID          string
	SystemBUID      string
}

type StartSessionResponse struct {
	EnableSessionSSL bool
	Request          string
	SessionID        string
}

func (l *LockdownConn) StartSession(cert *Certificate) (*StartSessionResponse, error) {
	if err := l.Send(startSessionRequest{
		Label:           Label,
		ProtocolVersion: "2",
		Request:         "StartSession",
		HostID:          cert.HostID,
		SystemBUID:      cert.SystemBUID,
	}); err != nil {
		return nil, err
	}

	body, err := l.Recv()
	if err != nil {
		return nil, err
	}

	var resp StartSessionResponse
	if _, err := plist.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	l.sessionId = resp.SessionID
	if resp.EnableSessionSSL {
		if err := l.Conn.EnableSessionSSL(cert); err != nil {
			return nil, err
		}
	}

	return &resp, nil
}

func (l *LockdownConn) GetValues() (map[string]interface{}, error) {
	req := valutRequest{
		Label:   Label,
		Key:     "",
		Request: "GetValue",
	}

	if err := l.Send(req); err != nil {
		return nil, err
	}

	bs, err := l.Recv()
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if _, err := plist.Unmarshal(bs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type StartServiceResponse struct {
	Port             uint16
	Request          string
	Service          string
	EnableServiceSSL bool
	Error            string
}

type startServiceRequest struct {
	Label   string
	Request string
	Service string
}

func (l *LockdownConn) StartService(name string) (*StartServiceResponse, error) {
	if err := l.Send(startServiceRequest{
		Label:   Label,
		Request: "StartService",
		Service: name,
	}); err != nil {
		return nil, err
	}

	body, err := l.Recv()
	if err != nil {
		return nil, err
	}

	var resp StartServiceResponse
	if _, err := plist.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, xerrors.Errorf("could not start service: %s with reason: %s.Have you mounted the Developer Image?", name, resp.Error)
	}

	return &resp, nil
}
