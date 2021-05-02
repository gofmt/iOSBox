package handlers

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
)

type DeviceInfo struct {
	Entry ios.DeviceEntry
	Value ios.AllValuesType
}

func CmdDeviceInfo(info *DeviceInfo) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "info",
		Help: "显示当前设备详细信息",
		Func: func(c *ishell.Context) {
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 0, 1, ' ', 0)
			value := info.Value
			_, _ = fmt.Fprintln(w, "- USBSerialNumber\t:\t"+info.Entry.Properties.SerialNumber)
			_, _ = fmt.Fprintln(w, "- DeviceName\t:\t"+value.DeviceName)
			_, _ = fmt.Fprintln(w, "- ProductName\t:\t"+value.ProductName)
			_, _ = fmt.Fprintln(w, "- ProductType\t:\t"+value.ProductType)
			_, _ = fmt.Fprintln(w, "- ProductVersion\t:\t"+value.ProductVersion)
			_, _ = fmt.Fprintln(w, "- CPUArchitecture\t:\t"+value.CPUArchitecture)
			_, _ = fmt.Fprintln(w, "- BuildVersion\t:\t"+value.BuildVersion)
			_, _ = fmt.Fprintln(w, "- SerialNumber\t:\t"+value.SerialNumber)
			_, _ = fmt.Fprintln(w, "- MLBSerialNumber\t:\t"+value.MLBSerialNumber)
			_, _ = fmt.Fprintln(w, "- BluetoothAddress\t:\t"+value.BluetoothAddress)
			_, _ = fmt.Fprintln(w, "- DeviceColor\t:\t"+value.DeviceColor)
			_, _ = fmt.Fprintln(w, "- FirmwareVersion\t:\t"+value.FirmwareVersion)
			_, _ = fmt.Fprintln(w, "- ActivationState\t:\t"+value.ActivationState)
			_, _ = fmt.Fprintln(w, "- HardwareModel\t:\t"+value.HardwareModel)
			_, _ = fmt.Fprintln(w, "- HardwarePlatform\t:\t"+value.HardwarePlatform)
			_ = w.Flush()
		},
	}
}
