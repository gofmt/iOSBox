package handlers

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/gookit/gcli/v3"
	"golang.org/x/xerrors"
)

var SystemRebootCommand = &gcli.Command{
	Name: "reboot",
	Desc: "重启当前设备，重启后需要重新越狱",
	Func: func(c *gcli.Command, args []string) error {
		device, err := ios.GetDevice("")
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}

		_ = diagnostics.Reboot(device)

		return nil
	},
}
