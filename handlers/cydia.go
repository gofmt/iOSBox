package handlers

import "github.com/gookit/gcli/v3"

var CydiaCommand = &gcli.Command{
	Name:    "cydia",
	Desc:    "Cydia 插件仓库",
	Aliases: []string{"ca"},
	Subs: []*gcli.Command{
		{
			Name:    "view",
			Desc:    "展示插件列表",
			Aliases: []string{"v"},
		},
		{
			Name:    "search",
			Desc:    "搜索插件",
			Aliases: []string{"s"},
		},
		{
			Name:    "install",
			Desc:    "安装插件",
			Aliases: []string{"i"},
		},
		{
			Name:    "uninstall",
			Desc:    "卸载插件",
			Aliases: []string{"u"},
		},
		{
			Name:    "upgrade",
			Desc:    "升级插件",
			Aliases: []string{"p"},
		},
		{
			Name:    "publish",
			Desc:    "发布插件",
			Aliases: []string{"pub"},
		},
	},
	Func: func(c *gcli.Command, args []string) error {
		c.ShowHelp()

		return nil
	},
}
