package handlers

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
)

func CmdProcessList(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "procs",
		Help: "显示进程列表",
		Func: func(c *ishell.Context) {
			dis, err := instruments.NewDeviceInfoService(entry)
			if err != nil {
				fmt.Println("连接服务错误", err)
				return
			}
			defer dis.Close()

			procList, err := dis.ProcessList()
			if err != nil {
				fmt.Println("获取进程列表错误:", err)
				return
			}

			fmt.Println("--------------------------------------------------------------")
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 0, 1, ' ', 0)
			for _, info := range procList {
				_, _ = fmt.Fprintln(w, fmt.Sprintf("PID: %d", info.Pid))
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Name: %s", info.Name))
				_, _ = fmt.Fprintln(w, fmt.Sprintf("Path: %s", info.RealAppName))
				_, _ = fmt.Fprintln(w, "--------------------------------------------------------------")
			}
			_ = w.Flush()
		},
	}
}

func CmdProcessKill(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "kill",
		Help: "输入PID结束进程",
		Func: func(c *ishell.Context) {
			conn, err := instruments.NewProcessControl(entry)
			if err != nil {
				fmt.Println("连接服务错误：", err)
				return
			}
			defer conn.Close()

			if len(c.Args) > 0 {
				pid, err := strconv.Atoi(c.Args[0])
				if err != nil {
					fmt.Println("PID错误：", err)
					return
				}

				if err := conn.KillProcess(uint64(pid)); err != nil {
					fmt.Printf("结束进程错误 [%d]：%v\n", pid, err)
				}

				return
			}

			fmt.Println("未输入进程pid")
		},
	}
}
