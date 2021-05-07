package idevice

import (
	"os"
	"os/signal"
	"testing"
)

func TestStartForward(t *testing.T) {
	device, err := GetDevice()
	if err != nil {
		t.Fatal(err)
	}

	service := NewForwardService(device)
	if err := service.Start(27042, 27042, func(msg string, nerr error) {
		if nerr != nil {
			t.Fatal(nerr)
		}

		t.Log(msg)
	}); err != nil {
		t.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
