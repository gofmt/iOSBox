package idevice

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

func TestMetchacall(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := NewInstrumentsService(device)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	msg, err := conn.channel.MethodCall("networkInformation")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(msg.PayloadHeader)

	t.Log(msg.Payload[0])
}
