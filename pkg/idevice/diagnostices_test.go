package idevice

import "testing"

func TestDiagnosticsService_AllValues(t *testing.T) {
	device, err := GetDevice()
	if err != nil {
		t.Fatal(err)
	}

	conn, err := NewDiagnosticsService(device)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	resp, err := conn.GetAllValues()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp)
}
