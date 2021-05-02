package handlers

import (
	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
)

func CmdRebootSystem(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "reboot",
		Help: "重启当前设备，重启后需要重新越狱",
		Func: func(c *ishell.Context) {
			_ = diagnostics.Reboot(entry)
		},
	}
}
