package handlers

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gofmt/iOSBox/pkg/idevice"

	"github.com/gookit/color"
	"github.com/gookit/gcli/v3"
	"golang.org/x/xerrors"
)

var SystemLogCommand = &gcli.Command{
	Name: "syslog",
	Desc: "打印系统日志",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "日志过滤字符串，支持过滤进程名或模块名")
	},
	Func: func(c *gcli.Command, args []string) error {
		device, err := idevice.GetDevice()
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}

		conn, err := idevice.NewSyslogService(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误: %w", err)
		}
		defer conn.Close()

		fgWhite := color.FgWhite.Render
		fgGreen := color.FgGreen.Render
		fgCyan := color.FgCyan.Render
		fgRed := color.FgRed.Render
		fgLiRed := color.FgLightRed.Render
		fgYellow := color.FgYellow.Render
		fgLiYellow := color.FgLightYellow.Render
		fgMagenta := color.FgMagenta.Render
		white := color.White.Render

		go func() {
			for {
				// Jun  3 18:45:44 iPhone wifid(WiFiPolicy)[51] <Notice>: Copy current network requested by "WirelessRadioMan"
				msg, err := conn.GetSyslog()
				if err != nil {
					fmt.Println("读取系统日志错误：", err)
					return
				}

				t, err := time.Parse(time.Stamp, msg.Time)
				if err != nil {
					panic(err)
				}

				// (kernel(AppleProxDriver)[0]) (进程(模块)[行号])
				if len(args) > 0 && !strings.Contains(msg.ProcInfo, args[0]) {
					continue
				}

				level := msg.Level
				body := msg.Body
				switch msg.Level {
				case "Notice":
					level = fgGreen(level)
				case "Error":
					level = fgRed(level)
					body = fgLiRed(body)
				case "Warning":
					level = fgYellow(level)
					body = fgLiYellow(body)
				case "Debug":
					level = fgMagenta(level)
				default:
					level = white(level)
				}

				fmt.Printf(
					"[%s](%s)[%s]: %s\n",
					fgWhite(t.Format("01-02 15:04:05")),
					// gray(msg.DeviceName),
					fgCyan(msg.ProcInfo),
					level,
					body,
				)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		return nil
	},
}
