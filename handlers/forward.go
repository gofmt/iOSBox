package handlers

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/gofmt/iOSBox/pkg/idevice"

	"github.com/gookit/gcli/v3"
)

var ForwardCommand = &gcli.Command{
	Name: "forward",
	Desc: "映射设备端口到本地端口(替代iproxy)",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "本机端口", true)
		c.AddArg("arg1", "设备端口", true)
	},
	Examples: "{$binName} {$cmd} 本机端口 设备端口",
	Func: func(c *gcli.Command, args []string) error {
		device, err := idevice.GetDevice()
		if err != nil {
			return err
		}

		localPort, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		remotePort, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		service := idevice.NewForwardService(device)
		err = service.Start(uint16(localPort), uint16(remotePort), func(s string, err error) {
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(s)
			}
		})
		if err != nil {
			return err
		}
		defer service.Close()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		return nil
	},
}
