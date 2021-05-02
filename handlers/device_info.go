package handlers

import (
	"fmt"
	"net"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
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
			ip := GetDeviceIPAddress(info.Entry)
			if ip != "" {
				_, _ = fmt.Fprintln(w, "- DeviceIPAddress\t:\t"+ip)
			}
			_ = w.Flush()
		},
	}
}

func GetDeviceIPAddress(entry ios.DeviceEntry) string {
	intf, err := ios.ConnectToService(entry, "com.apple.pcapd")
	if err != nil {
		fmt.Println("连接服务错误：", err)
		return ""
	}
	defer intf.Close()

	pListCodec := ios.NewPlistCodec()
	for {
		bs, err := pListCodec.Decode(intf.Reader())
		if err != nil {
			fmt.Println("读取网络封包错误：", err)
			return ""
		}
		_, err = plist.Unmarshal(bs, &bs)
		if err != nil {
			fmt.Println("iOS包反系列化错误: ", err)
			return ""
		}

		// 109 iOS封包的头部大小，剩余的是TCP包体，其中头部20字节为IP层结构，取IP头4字节的IP地址
		ipbytes := bs[109+12 : 109+12+4]
		DeviceIP = net.IP(ipbytes).String()
		if strings.HasPrefix(DeviceIP, "192.168.") { // TODO 可能会有问题,需要反查一下hostname
			break
		}
	}
	intf.Close()

	return DeviceIP
}
