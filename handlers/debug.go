package handlers

import "github.com/gookit/gcli/v3"

var DebugCommand  = &gcli.Command{
	Name: "dbgserver",
	Desc: "启动设备上的 debug-server",
	Func: func(c *gcli.Command, args []string) error {

		return nil
	},
}

var LLDBCommand = &gcli.Command{
	Name: "lldb",
	Desc: "多窗口 lldb 调试",
	Func: func(c *gcli.Command, args []string) error {
		return nil
	},
}