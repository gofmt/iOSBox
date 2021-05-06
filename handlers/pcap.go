package handlers

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"

	"iOSBox/pkg/idevice"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gookit/gcli/v3"
)

var PcapCommand = &gcli.Command{
	Name: "pcap",
	Desc: "网络抓包",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "PCAP文件保存路径", true)
		c.AddArg("arg1", "进程名称")
	},
	Func: func(c *gcli.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		device, err := ios.GetDevice("")
		if err != nil {
			return err
		}

		f, err := os.OpenFile(args[0], os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		procName := ""
		if len(args) == 2 {
			procName = args[1]
		}

		go func() {
			if err := idevice.StartPcapService(ctx, device, procName, f, func(bs []byte) {
				fmt.Println(hex.Dump(bs))
			}); err != nil {
				fmt.Println("抓包错误：", err)
				os.Exit(-1)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		cancel()

		return nil
	},
}
