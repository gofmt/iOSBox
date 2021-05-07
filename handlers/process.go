package handlers

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gookit/gcli/v3"
)

var ProcessListCommand = &gcli.Command{
	Name:    "procs",
	Desc:    "显示当前设备进程列表",
	Aliases: []string{"ps"},
	Func: func(c *gcli.Command, args []string) error {
		result, err := shellRun("ps -eec")
		if err != nil {
			return err
		}

		if len(result) == 0 {
			return nil
		}

		results := bytes.Split(result, []byte("\n"))
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 0, 4, ' ', 0)

		for _, res := range results[1:] {
			ss := bytes.Fields(res)
			if len(ss) > 0 {
				_, _ = fmt.Fprintln(w, string(ss[0])+"\t"+string(ss[3]))
			}
		}

		_ = w.Flush()

		return nil
	},
}

// var AppLaunchCommand = &gcli.Command{
// 	Name:    "launch",
// 	Desc:    "启动应用",
// 	Aliases: []string{"l"},
// 	Config: func(c *gcli.Command) {
// 		c.AddArg("arg0", "应用BundleID", true)
// 	},
// 	Examples: "{$binName} {$cmd} con.xxx.xxx",
// 	Func: func(c *gcli.Command, args []string) error {
// 		if len(args) == 0 {
// 			return xerrors.Errorf("未传入应用BundleID")
// 		}
//
// 		device, err := ios.GetDevice("")
// 		if err != nil {
// 			return xerrors.Errorf("连接iOS设备错误: %w", err)
// 		}
//
// 		conn, err := instruments.NewProcessControl(device)
// 		if err != nil {
// 			return xerrors.Errorf("连接服务错误：%w", err)
// 		}
// 		defer conn.Close()
//
// 		bundleId := args[0]
// 		pid, err := conn.LaunchApp(bundleId)
// 		if err != nil {
// 			return xerrors.Errorf("启动应用错误：%w", err)
// 		}
//
// 		fmt.Printf("应用 %s 已启动，PID = %d\n", bundleId, pid)
//
// 		return nil
// 	},
// }

var ProcessKillCommand = &gcli.Command{
	Name:     "kill",
	Desc:     "结束进程",
	Aliases:  []string{"k"},
	Examples: "{$binName} {$cmd} SpringBoard",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "PID或进程名", true)
	},
	Func: func(c *gcli.Command, args []string) error {
		result, err := shellRun("ps -eec")
		if err != nil {
			return err
		}

		results := bytes.Split(result, []byte("\n"))
		if len(results) == 0 {
			return nil
		}

		for _, line := range results[1:] {
			fields := bytes.Fields(line)
			if len(fields) == 0 {
				continue
			}

			if string(fields[0]) == args[0] || string(fields[3]) == args[0] {
				_, _ = shellRun("kill " + string(fields[0]))
				break
			}
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
