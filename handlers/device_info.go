package handlers

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gofmt/iOSBox/pkg/idevice"

	"github.com/gookit/gcli/v3"
	"golang.org/x/xerrors"
)

var DeviceInfoCommand = &gcli.Command{
	Name:    "info",
	Desc:    "显示当前设备信息",
	Aliases: []string{"in"},
	Func: func(c *gcli.Command, args []string) error {
		device, err := idevice.GetDevice()
		if err != nil {
			return xerrors.Errorf("获取iOS设备错误: %w", err)
		}

		lockdown, err := idevice.ConnectLockdownWithSession(device)
		if err != nil {
			return err
		}

		ret, err := lockdown.GetValues()
		if err != nil {
			return xerrors.Errorf("获取设备信息错误：%w", err)
		}

		info := ret["Value"].(map[string]interface{})

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 0, 1, ' ', 0)

		_, _ = fmt.Fprintln(w, "- UDID\t: "+info["UniqueDeviceID"].(string))
		_, _ = fmt.Fprintln(w, "- DeviceName\t: "+info["DeviceName"].(string))
		_, _ = fmt.Fprintln(w, "- ProductName\t: "+info["ProductName"].(string))
		_, _ = fmt.Fprintln(w, "- ProductType\t: "+info["ProductType"].(string))
		_, _ = fmt.Fprintln(w, "- ProductVersion\t: "+info["ProductVersion"].(string))
		_, _ = fmt.Fprintln(w, "- CPUArchitecture\t: "+info["CPUArchitecture"].(string))
		_, _ = fmt.Fprintln(w, "- BuildVersion\t: "+info["BuildVersion"].(string))
		_, _ = fmt.Fprintln(w, "- SerialNumber\t: "+info["SerialNumber"].(string))
		_, _ = fmt.Fprintln(w, "- MLBSerialNumber\t: "+info["MLBSerialNumber"].(string))
		_, _ = fmt.Fprintln(w, "- BluetoothAddress\t: "+info["BluetoothAddress"].(string))
		_, _ = fmt.Fprintln(w, "- WiFiAddress\t: "+info["WiFiAddress"].(string))
		_, _ = fmt.Fprintln(w, "- EthernetAddress\t: "+info["EthernetAddress"].(string))
		_, _ = fmt.Fprintln(w, "- DeviceColor\t: "+info["DeviceColor"].(string))
		_, _ = fmt.Fprintln(w, "- FirmwareVersion\t: "+info["FirmwareVersion"].(string))
		_, _ = fmt.Fprintln(w, "- ActivationState\t: "+info["ActivationState"].(string))
		_, _ = fmt.Fprintln(w, "- HardwareModel\t: "+info["HardwareModel"].(string))
		_, _ = fmt.Fprintln(w, "- HardwarePlatform\t: "+info["HardwarePlatform"].(string))
		_, _ = fmt.Fprintln(w, "- UniqueChipID\t: "+fmt.Sprintf("%d", info["UniqueChipID"]))
		_, _ = fmt.Fprintln(w, "- WirelessBoardSerialNumber\t: "+info["WirelessBoardSerialNumber"].(string))

		_ = w.Flush()

		return nil
	},
}
