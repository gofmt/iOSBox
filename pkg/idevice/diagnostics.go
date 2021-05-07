package idevice

import (
	"golang.org/x/xerrors"
	"howett.net/plist"
)

type allDiagnosticsResponse struct {
	Diagnostics Diagnostics
	Status      string
}

type Diagnostics struct {
	GasGauge GasGauge
	HDMI     HDMI
	NAND     NAND
	WiFi     WiFi
}

type WiFi struct {
	Active string
	Status string
}

type NAND struct {
	Status string
}

type HDMI struct {
	Connection string
	Status     string
}

type GasGauge struct {
	CycleCount         uint64
	DesignCapacity     uint64
	FullChargeCapacity uint64
	Status             string
}

type diagnosticsRequest struct {
	Request string
}

type diagnosticsStatus struct {
	Status string
}

type rebootRequest struct {
	Request           string
	WaitForDisconnect bool
	DisplayPass       bool
	DisplayFail       bool
}

type DiagnosticsService struct {
	conn IConn
}

func NewDiagnosticsService(entry *DeviceEntry) (*DiagnosticsService, error) {
	conn, err := ConnectToService(entry, "com.apple.mobile.diagnostics_relay")
	if err != nil {
		return nil, err
	}

	return &DiagnosticsService{conn: conn}, nil
}

func (d *DiagnosticsService) GetAllValues() (*allDiagnosticsResponse, error) {
	req := diagnosticsRequest{"All"}
	bs, err := d.conn.Encode(req)
	if err != nil {
		return nil, err
	}

	if err := d.conn.Write(bs); err != nil {
		return nil, err
	}

	body, err := d.conn.Decode(d.conn.Reader())
	if err != nil {
		return nil, err
	}

	var resp allDiagnosticsResponse
	if _, err := plist.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (d *DiagnosticsService) Reboot() error {
	req := rebootRequest{
		Request:           "Restart",
		WaitForDisconnect: true,
		DisplayFail:       true,
		DisplayPass:       true,
	}
	bs, err := d.conn.Encode(req)
	if err != nil {
		return err
	}

	if err := d.conn.Write(bs); err != nil {
		return err
	}

	body, err := d.conn.Decode(d.conn.Reader())
	if err != nil {
		return err
	}

	var resp map[string]interface{}
	if _, err := plist.Unmarshal(body, &resp); err != nil {
		return err
	}

	if val, ok := resp["Status"].(string); ok {
		if val == "Success" {
			return nil
		}
	}

	return xerrors.Errorf("Could not reboot, response: %+v", resp)
}

func (d *DiagnosticsService) Close() {
	req := diagnosticsRequest{"Goodbye"}
	bs, err := d.conn.Encode(req)
	if err != nil {
		return
	}

	if err := d.conn.Write(bs); err != nil {
		return
	}

	_, _ = d.conn.Decode(d.conn.Reader())
	d.conn.Close()
}
