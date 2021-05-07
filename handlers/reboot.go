package handlers

import (
	"iOSBox/pkg/idevice"

	"github.com/gookit/gcli/v3"
	"golang.org/x/xerrors"
)

var SystemRebootCommand = &gcli.Command{
	Name: "reboot",
	Desc: "重启当前设备，重启后需要重新越狱",
	Func: func(c *gcli.Command, args []string) error {
		device, err := idevice.GetDevice()
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}

		conn, err := idevice.NewDiagnosticsService(device)
		if err != nil {
			return err
		}
		defer conn.Close()

		_ = conn.Reboot()

		return nil
	},
}
