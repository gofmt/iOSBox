package idevice

import "testing"

func TestLockdownConn_GetValues(t *testing.T) {
	device, err := GetDevice()
	if err != nil {
		t.Fatal(err)
	}

	lockdown, err := ConnectLockdownWithSession(device)
	if err != nil {
		t.Fatal(err)
	}
	defer lockdown.Close()

	resp, err := lockdown.GetValues()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp)
}
