package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"iOSBox/pkg/frida"

	"github.com/gookit/gcli/v3"
	"golang.org/x/xerrors"
)

var RunFridaScriptCommand = &gcli.Command{
	Name: "script",
	Desc: "执行frida脚本",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "应用BundleID", true)
		c.AddArg("arg1", "脚本路径", true)
	},
	Examples: "{$binName} {$cmd} testdata/example.js",
	Func: func(c *gcli.Command, args []string) error {
		bs, err := ioutil.ReadFile(args[1])
		if err != nil {
			return xerrors.Errorf("读取脚本文件错误：%w", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			if err := frida.StartFrida(ctx, os.Stdout, args[0], string(bs)); err != nil {
				fmt.Println("执行脚本错误：", err)
				os.Exit(-1)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
		<-quit

		cancel()

		return nil
	},
}
