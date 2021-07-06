package main

import (
	"fmt"
	"os"

	"github.com/gofmt/iOSBox/handlers"
	"github.com/gookit/gcli/v3"
)

func main() {
	gcli.DefaultVerb = gcli.VerbQuiet

	app := gcli.NewApp(func(app *gcli.App) {
		app.Version = "v0.1.1-alpha"
		app.Desc = "iOSBox"
		app.ExitOnEnd = false
		app.On(gcli.EvtAppInit, func(data ...interface{}) (stop bool) {
			// fmt.Println("init app")
			return false
		})
	})

	app.On(gcli.EvtAppRunError, func(data ...interface{}) (stop bool) {
		fmt.Println(data[1])
		return false
	})

	app.On(gcli.EvtCmdNotFound, func(data ...interface{}) (stop bool) {
		return false
	})

	app.On(gcli.EvtAppCmdNotFound, func(data ...interface{}) (stop bool) {
		return false
	})

	app.On(gcli.EvtCmdRunError, func(data ...interface{}) (stop bool) {
		return false
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
		handlers.DebugCommand,
		handlers.LLDBCommand,
		handlers.FridaCommand,
	)

	os.Exit(app.Run(nil))
}
