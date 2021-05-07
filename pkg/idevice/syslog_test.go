package idevice

import (
	"testing"
)

func TestSyslogService_GetLog(t *testing.T) {
	device, err := GetDevice()
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewSyslogService(device)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()

	for {
		line, err := service.GetSyslog()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(line)
	}
}
