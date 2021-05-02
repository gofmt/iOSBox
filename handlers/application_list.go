package handlers

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
)

func CmdApplicationList(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "apps",
		Help: "当前设备已安装应用列表",
		Func: func(c *ishell.Context) {
			conn, err := installationproxy.New(entry)
			if err != nil {
				fmt.Println("连接安装服务错误：", err)
				return
			}
			defer conn.Close()

			userAppList, err := conn.BrowseUserApps()
			if err != nil {
				fmt.Println("获取用户应用列表错误：", err)
				return
			}

			sysAppList, err := conn.BrowseSystemApps()
			if err != nil {
				fmt.Println("获取系统应用列表错误：", err)
				return
			}

			appList := make([]installationproxy.AppInfo, 0)
			appList = append(appList, userAppList...)
			appList = append(appList, sysAppList...)

			fmt.Println("--------------------------------------------------------------")
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 0, 1, ' ', 0)
			for _, info := range appList {
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Name\t: %s", info.CFBundleDisplayName))
				_, _ = fmt.Fprintln(w, fmt.Sprintf("BundleId\t: %s", info.CFBundleIdentifier))
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Version\t: %s", info.CFBundleShortVersionString))
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Executable\t: %s", info.CFBundleExecutable))
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Container\t: %s", info.Container))
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Path\t: %s", info.Path))
				_, _ = fmt.Fprintln(w, "--------------------------------------------------------------")
			}
			_ = w.Flush()
		},
	}
}
