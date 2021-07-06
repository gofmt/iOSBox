package handlers

import "github.com/gookit/gcli/v3"

var FridaCommand = &gcli.Command{
	Name:    "frida",
	Desc:    "Frida 脚本仓库",
	Aliases: []string{"fa"},
	Subs: []*gcli.Command{
		{
			Name: "list",
			Desc: "脚本列表",
		},
		{
			Name: "search",
			Desc: "搜索脚本",
		},
		{
			Name: "get",
			Desc: "下载脚本",
		},
	},
	Func: func(c *gcli.Command, args []string) error {

		return nil
	},
}
