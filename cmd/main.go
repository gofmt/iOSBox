package main

import (
	"fmt"
	"os"
	"strconv"

	"iOSBox/handlers"

	"github.com/abiosoft/ishell"
	"github.com/abiosoft/readline"
	"github.com/danielpaulus/go-ios/ios"
)

func main() {
	deviceList, err := ios.ListDevices()
	if err != nil {
		fmt.Println("获取iOS设备列表错误：", err)
		os.Exit(-1)
	}

	devices := make([]handlers.DeviceInfo, 0)
	for _, entry := range deviceList.DeviceList {
		info := handlers.DeviceInfo{Entry: entry}
		resp, err := ios.GetValues(entry)
		if err != nil {
			fmt.Println("获取iOS设备信息错误：", err)
			os.Exit(-1)
		}

		info.Value = resp.Value
		devices = append(devices, info)
	}

	fmt.Println("设备列表:")
	fmt.Println("------------------------------------------------------------------------------------------------")

	for i, device := range devices {
		fmt.Println(i, "\t|", device.Entry.Properties.SerialNumber, "|", device.Value.DeviceName, "|",
			device.Value.SerialNumber, "|", device.Value.MLBSerialNumber, "|")
	}

	fmt.Println("------------------------------------------------------------------------------------------------")
	fmt.Println("输入设备编号：")
	var input string
	_, err = fmt.Scan(&input)
	if err != nil {
		fmt.Println("扫描终端输入错误：", err)
		os.Exit(-1)
	}
	idx, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("输入的设备编号错误：", err)
		os.Exit(-1)
	}
	if idx > len(devices)-1 {
		fmt.Println("没有这个设备编号：", idx)
		os.Exit(-1)
	}

	fmt.Println("================================================================================================")

	deviceName := devices[idx].Value.DeviceName
	shell := ishell.NewWithConfig(&readline.Config{Prompt: deviceName + "> "})
	defer shell.Close()
	shell.AddCmd(handlers.CmdDeviceInfo(&devices[idx]))
	shell.AddCmd(handlers.CmdApplicationList(devices[idx].Entry))
	fmt.Println(shell.HelpText())
	shell.Run()
}
