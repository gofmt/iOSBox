package handlers

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/gookit/gcli/v3"
	"golang.org/x/xerrors"
)

var ProcessListCommand = &gcli.Command{
	Name:    "procs",
	Desc:    "显示当前设备进程列表",
	Aliases: []string{"ps"},
	Func: func(c *gcli.Command, args []string) error {
		device, err := ios.GetDevice("")
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}

		conn, err := instruments.NewDeviceInfoService(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误: %w", err)
		}
		defer conn.Close()

		procList, err := conn.ProcessList()
		if err != nil {
			return xerrors.Errorf("获取进程列表错误：%w", err)
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

		return nil
	},
}

var AppLaunchCommand = &gcli.Command{
	Name:    "launch",
	Desc:    "启动应用",
	Aliases: []string{"l"},
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "应用BundleID", true)
	},
	Examples: "{$binName} {$cmd} con.xxx.xxx",
	Func: func(c *gcli.Command, args []string) error {
		if len(args) == 0 {
			return xerrors.Errorf("未传入应用BundleID")
		}

		device, err := ios.GetDevice("")
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}

		conn, err := instruments.NewProcessControl(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误：%w", err)
		}
		defer conn.Close()

		bundleId := args[0]
		pid, err := conn.LaunchApp(bundleId)
		if err != nil {
			return xerrors.Errorf("启动应用错误：%w", err)
		}

		fmt.Printf("应用 %s 已启动，PID = %d\n", bundleId, pid)

		return nil
	},
}

var ProcessKillCommand = &gcli.Command{
	Name:     "kill",
	Desc:     "结束进程",
	Aliases:  []string{"k"},
	Examples: "{$binName} {$cmd} SpringBoard",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "PID或进程名", true)
	},
	Func: func(c *gcli.Command, args []string) error {
		if len(args) == 0 {
			return xerrors.Errorf("未传入PID或进程名")
		}

		device, err := ios.GetDevice("")
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}

		conn, err := instruments.NewProcessControl(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误: %w", err)
		}
		defer conn.Close()

		pid, err := strconv.Atoi(args[0])
		if err != nil {
			dis, err := instruments.NewDeviceInfoService(device)
			if err != nil {
				return xerrors.Errorf("连接服务错误: %w", err)
			}
			defer dis.Close()

			procList, err := dis.ProcessList()
			if err != nil {
				return xerrors.Errorf("获取进程列表错误: %w", err)
			}

			for _, info := range procList {
				if info.Name == args[0] {
					pid = int(info.Pid)
				}
			}
		}

		if pid == 0 {
			return xerrors.New("进程PID不能为：0")
		}

		if err := conn.KillProcess(uint64(pid)); err != nil {
			return xerrors.Errorf("[%d]结束进程错误: %w", pid, err)
		}

		return nil
	},
}

var ReSpringBoardCommand = &gcli.Command{
	Name: "respring",
	Desc: "重启 SpringBoard",
	Func: func(c *gcli.Command, args []string) error {
		return c.App().Exec("kill", []string{"SpringBoard"})
	},
}
