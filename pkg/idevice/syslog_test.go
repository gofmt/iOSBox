package idevice

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

func TestSyslogService_GetLog(t *testing.T) {
	device, err := ios.GetDevice("")
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
