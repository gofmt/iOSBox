package idevice

import (
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

type AllValuesType struct {
	ActivationState                             string
	ActivationStateAcknowledged                 bool
	BasebandActivationTicketVersion             string
	BasebandCertID                              int `plist:"BasebandCertId"`
	BasebandChipID                              int
	BasebandKeyHashInformation                  BasebandKeyHashInformationType
	BasebandMasterKeyHash                       string
	BasebandRegionSKU                           []byte
	BasebandSerialNumber                        []byte
	BasebandStatus                              string
	BasebandVersion                             string
	BluetoothAddress                            string
	BoardID                                     int `plist:"BoardId"`
	BrickState                                  bool
	BuildVersion                                string
	CPUArchitecture                             string
	CarrierBundleInfoArray                      []interface{}
	CertID                                      int
	ChipID                                      int
	ChipSerialNo                                []byte
	DeviceClass                                 string
	DeviceColor                                 string
	DeviceName                                  string
	DieID                                       int
	EthernetAddress                             string
	FirmwareVersion                             string
	FusingStatus                                int
	HardwareModel                               string
	HardwarePlatform                            string
	HasSiDP                                     bool
	HostAttached                                bool
	InternationalMobileEquipmentIdentity        string
	MLBSerialNumber                             string
	MobileEquipmentIdentifier                   string
	MobileSubscriberCountryCode                 string
	MobileSubscriberNetworkCode                 string
	ModelNumber                                 string
	NonVolatileRAM                              NonVolatileRAMType
	PartitionType                               string
	PasswordProtected                           bool
	PkHash                                      []byte
	ProductName                                 string
	ProductType                                 string
	ProductVersion                              string
	ProductionSOC                               bool
	ProtocolVersion                             string
	ProximitySensorCalibration                  []byte
	RegionInfo                                  string
	SBLockdownEverRegisteredKey                 bool
	SIMStatus                                   string
	SIMTrayStatus                               string
	SerialNumber                                string
	SoftwareBehavior                            []byte
	SoftwareBundleVersion                       string
	SupportedDeviceFamilies                     []int
	TelephonyCapability                         bool
	TimeIntervalSince1970                       float64
	TimeZone                                    string
	TimeZoneOffsetFromUTC                       float64
	TrustedHostAttached                         bool
	UniqueChipID                                uint64
	UniqueDeviceID                              string
	UseRaptorCerts                              bool
	Uses24HourClock                             bool
	WiFiAddress                                 string
	WirelessBoardSerialNumber                   string
	KCTPostponementInfoPRIVersion               string `plist:"kCTPostponementInfoPRIVersion"`
	KCTPostponementInfoPRLName                  int    `plist:"kCTPostponementInfoPRLName"`
	KCTPostponementInfoServiceProvisioningState bool   `plist:"kCTPostponementInfoServiceProvisioningState"`
	KCTPostponementStatus                       string `plist:"kCTPostponementStatus"`
}

type GetAllValuesResponse struct {
	Request string
	Value   AllValuesType
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

func (l *LockdownConn) GetValues() (*GetAllValuesResponse, error) {
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

	var resp GetAllValuesResponse
	if _, err := plist.Unmarshal(bs, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
