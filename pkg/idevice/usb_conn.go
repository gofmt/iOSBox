package idevice

import (
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/xerrors"
	"howett.net/plist"
)

const (
	Label               = "idevice.ios.control"
	ProgName            = "idevice-usbmux"
	ClientVersionString = "idevice-usbmux-0.0.1"
)

type USBHeader struct {
	Length  uint32
	Version uint32
	Request uint32
	Tag     uint32
}

type USBMessage struct {
	Header  USBHeader
	Payload []byte
}

type USBResponse struct {
	MessageType string
	Number      uint32
}

type DeviceProperties struct {
	ConnectionSpeed int
	ConnectionType  string
	DeviceID        int
	LocationID      int
	ProductID       int
	SerialNumber    string
}

type DeviceEntry struct {
	DeviceID    int
	MessageType string
	Properties  DeviceProperties
}

type ListDevicesMessage struct {
	MessageType         string
	ProgName            string
	ClientVersionString string
}

type USBConn struct {
	Conn IConn
	tag  uint32
}

func NewUSBConn() (*USBConn, error) {
	conn, err := NewConn()
	if err != nil {
		return nil, err
	}

	return &USBConn{Conn: conn}, nil
}

func (u *USBConn) Close() {
	u.Conn.Close()
}

type connectMessage struct {
	BundleID            string
	ClientVersionString string
	MessageType         string
	ProgName            string
	LibUSBMuxVersion    uint32 `plist:"kLibUSBMuxVersion"`
	DeviceID            uint32
	PortNumber          uint16
}

func (u *USBConn) Connect(deviceId int, port uint16) error {
	if err := u.Send(connectMessage{
		BundleID:            Label,
		ClientVersionString: ClientVersionString,
		MessageType:         "Connect",
		ProgName:            ProgName,
		LibUSBMuxVersion:    3,
		DeviceID:            uint32(deviceId),
		PortNumber:          port,
	}); err != nil {
		return err
	}

	msg, err := u.Recv()
	if err != nil {
		return err
	}

	var resp USBResponse
	if _, err := plist.Unmarshal(msg.Payload, &resp); err != nil {
		return err
	}
	if resp.Number != 0 {
		return xerrors.Errorf("failed connecting to service, error code: %d", resp.Number)
	}

	return nil
}

func (u *USBConn) SendWithMessage(msg *USBMessage) error {
	if err := binary.Write(u.Conn.Writer(), binary.LittleEndian, msg.Header); err != nil {
		return err
	}

	return binary.Write(u.Conn.Writer(), binary.LittleEndian, msg.Payload)
}

func (u *USBConn) Send(msg interface{}) error {
	u.tag++

	bs, err := plist.Marshal(msg, plist.XMLFormat)
	if err != nil {
		return err
	}

	header := USBHeader{
		Length:  16 + uint32(len(bs)),
		Version: 1,
		Request: 8,
		Tag:     u.tag,
	}
	if err := binary.Write(u.Conn.Writer(), binary.LittleEndian, header); err != nil {
		return err
	}

	return binary.Write(u.Conn.Writer(), binary.LittleEndian, bs)
}

func (u *USBConn) Recv() (*USBMessage, error) {
	var header USBHeader
	if err := binary.Read(u.Conn.Reader(), binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	payload := make([]byte, header.Length-16)
	if _, err := io.ReadFull(u.Conn.Reader(), payload); err != nil {
		return nil, err
	}

	return &USBMessage{Header: header, Payload: payload}, nil
}

type Certificate struct {
	HostID            string
	SystemBUID        string
	HostCertificate   []byte
	HostPrivateKey    []byte
	DeviceCertificate []byte
	EscrowBag         []byte
	WiFiMACAddress    string
	RootCertificate   []byte
	RootPrivateKey    []byte
}

type ReadPairRecordRequest struct {
	BundleID            string
	ClientVersionString string
	MessageType         string
	ProgName            string
	LibUSBMuxVersion    uint32 `plist:"kLibUSBMuxVersion"`
	PairRecordID        string
}

type PairRecordData struct {
	PairRecordData []byte
}

func (u *USBConn) GetCertificate(udid string) (*Certificate, error) {
	if err := u.Send(ReadPairRecordRequest{
		BundleID:            Label,
		ClientVersionString: ClientVersionString,
		MessageType:         "ReadPairRecord",
		ProgName:            ProgName,
		LibUSBMuxVersion:    3,
		PairRecordID:        udid,
	}); err != nil {
		return nil, err
	}

	msg, err := u.Recv()
	if err != nil {
		return nil, err
	}

	var data PairRecordData
	if _, err := plist.Unmarshal(msg.Payload, &data); err != nil {
		return nil, err
	}

	var cert Certificate
	if _, err := plist.Unmarshal(data.PairRecordData, &cert); err != nil {
		return nil, err
	}

	return &cert, nil
}

func (u *USBConn) ListDevices() ([]DeviceEntry, error) {
	if err := u.Send(ListDevicesMessage{
		MessageType:         "ListDevices",
		ProgName:            ProgName,
		ClientVersionString: ClientVersionString,
	}); err != nil {
		return nil, err
	}

	msg, err := u.Recv()
	if err != nil {
		return nil, err
	}

	var deviceList = struct {
		DeviceList []DeviceEntry
	}{}

	if _, err := plist.Unmarshal(msg.Payload, &deviceList); err != nil {
		return nil, err
	}

	return deviceList.DeviceList, nil
}

func (u *USBConn) ConnectLockdown(devicdId int) (*LockdownConn, error) {
	if err := u.Connect(devicdId, 32498); err != nil {
		return nil, err
	}

	return &LockdownConn{Conn: u.Conn}, nil
}

func ConnectLockdownWithSession(entry *DeviceEntry) (*LockdownConn, error) {
	conn, err := NewUSBConn()
	if err != nil {
		return nil, err
	}
	// defer conn.Close()

	cert, err := conn.GetCertificate(entry.Properties.SerialNumber)
	if err != nil {
		return nil, err
	}

	lockdown, err := conn.ConnectLockdown(entry.DeviceID)
	if err != nil {
		return nil, err
	}

	_, err = lockdown.StartSession(cert)
	if err != nil {
		return nil, err
	}

	return lockdown, nil
}

func GetDevice(udid ...string) (device *DeviceEntry, err error) {
	conn, err := NewUSBConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	list, err := conn.ListDevices()
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		device = &list[0]
		if len(udid) > 0 {
			for _, entry := range list {
				if udid[0] == entry.Properties.SerialNumber {
					device = &entry
					break
				}
			}
		}

		return
	}

	return nil, errors.New("没有连接任何iOS设备")
}

func Ntohs(port uint16) uint16 {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	return binary.LittleEndian.Uint16(buf)
}
