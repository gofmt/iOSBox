package handlers

import (
	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
)

func CmdPcap(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "pcap",
		Help: "网络抓包",
		Func: func(c *ishell.Context) {

		},
	}
}
