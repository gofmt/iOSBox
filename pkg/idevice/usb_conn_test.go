package idevice

import "testing"

func TestUSBConn_ListDevices(t *testing.T) {
	device, err := GetDevice()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := NewUSBConn()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	cert, err := conn.GetCertificate(device.Properties.SerialNumber)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", cert)
}
