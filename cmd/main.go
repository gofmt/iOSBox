package main

import (
	"fmt"

	"iOSBox/handlers"

	"github.com/gookit/gcli/v3"
	"github.com/sirupsen/logrus"
)

func main() {
	// logrus.SetOutput(io.Discard)
	logrus.SetLevel(logrus.DebugLevel)

	gcli.AppHelpTemplate = `{{.Desc}} (版本: <info>{{.Version}}</>)
-----------------------------------------------------{{range $cmdName, $c := .Cs}}
  <info>{{$c.Name | paddingName }}</> {{$c.HelpDesc}}{{if $c.Aliases}} (别名: <green>{{ join $c.Aliases ","}}</>){{end}}{{end}}
  <info>{{ paddingName "help" }}</> 显示帮助信息

使用 "<cyan>{$binName} COMMAND -h</>" 查看命令的其他帮助信息
`

	app := gcli.NewApp(func(app *gcli.App) {
		app.Version = "v0.1.1-alpha"
		app.Desc = "iOSBox"
		app.On(gcli.EvtAppInit, func(data ...interface{}) (stop bool) {
			// fmt.Println("init app")
			return false
		})
	})

	app.On(gcli.EvtAppRunError, func(data ...interface{}) (stop bool) {
		fmt.Println(data[1])
		return true
	})

	app.Add(
		handlers.DeviceInfoCommand,
		handlers.AppListCommand,
		handlers.AppInstallCommand,
		handlers.AppUninstallCommand,
		handlers.ProcessListCommand,
		handlers.ProcessKillCommand,
		handlers.SystemRebootCommand,
		handlers.SystemLogCommand,
		handlers.ShellCommand,
		handlers.LdrestartCommand,
		handlers.RunCommand,
		handlers.ReSpringBoardCommand,
		handlers.ForwardCommand,
		handlers.SCPCommand,
		handlers.DonationCommand,
		handlers.PcapCommand,
	)

	app.Run(nil)
}
