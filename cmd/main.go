package main

import (
	"fmt"

	"iOSBox/handlers"

	"github.com/gookit/gcli/v3"
)

func main() {

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
		handlers.PcapCommand,
	)

	app.Run(nil)
}
