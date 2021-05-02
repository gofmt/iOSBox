package handlers

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/syslog"
)

func CmdSyslog(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "syslog",
		Help: "显示系统日志",
		Func: func(c *ishell.Context) {
			conn, err := syslog.New(entry)
			if err != nil {
				fmt.Println("连接服务错误：", err)
				return
			}
			defer conn.Close()

			go func() {
				for {
					line, err := conn.ReadLogMessage()
					if err != nil {
						fmt.Println("读取系统日志错误：", err)
						return
					}

					fmt.Print(line)
				}
			}()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			<-quit
		},
	}
}
