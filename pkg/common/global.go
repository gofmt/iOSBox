package common

import (
	"fmt"
	"net"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
)

var (
	Version     = "0.1.1-alpha"
	DeviceIP    = ""
	SSHUserName = "root"
	SSHPassword = "alpine"
)

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
