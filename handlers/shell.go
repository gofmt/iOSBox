package handlers

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"howett.net/plist"
)

func CmdShell(deviceInfo *DeviceInfo) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "ssh",
		Help: "创建SSH交互式环境，需要越狱",
		Func: func(c *ishell.Context) {
			intf, err := ios.ConnectToService(deviceInfo.Entry, "com.apple.pcapd")
			if err != nil {
				fmt.Println("连接服务错误：", err)
				return
			}
			defer intf.Close()

			pListCodec := ios.NewPlistCodec()
			var ip string
			for {
				bs, err := pListCodec.Decode(intf.Reader())
				if err != nil {
					fmt.Println("读取网络封包错误：", err)
					return
				}
				_, err = plist.Unmarshal(bs, &bs)
				if err != nil {
					fmt.Println("iOS包反系列化错误: ", err)
					return
				}

				// 109 iOS封包的头部大小，剩余的是TCP包体，其中头部20字节为IP层结构，取IP头4字节的IP地址
				ipbytes := bs[109+12 : 109+12+4]
				ip = net.IP(ipbytes).String()
				if strings.HasPrefix(ip, "192.168.") { // TODO 可能会有问题,需要反查一下hostname
					break
				}
			}
			intf.Close()
			if ip == "" {
				fmt.Println("未获取到设备IP地址")
				return
			}

			cfg := ssh.ClientConfig{
				User: "root",
				Auth: []ssh.AuthMethod{
					ssh.Password("alpine"),
				},
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
					return nil
				},
				Timeout: 10 * time.Second,
			}
			cli, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", ip), &cfg)
			if err != nil {
				fmt.Println("连接SSH服务器错误：", err)
				return
			}
			defer func(cli *ssh.Client) {
				_ = cli.Close()
			}(cli)

			session, err := cli.NewSession()
			if err != nil {
				fmt.Println("创建SSH会话错误：", err)
				return
			}
			defer func(session *ssh.Session) {
				_ = session.Close()
			}(session)

			fd := int(os.Stdin.Fd())
			oldState, err := terminal.MakeRaw(fd)
			if err != nil {
				fmt.Println("创建SSH终端错误：", err)
				return
			}
			defer func(fd int, oldState *terminal.State) {
				_ = terminal.Restore(fd, oldState)
			}(fd, oldState)

			session.Stdout = os.Stdout
			session.Stdin = os.Stdin
			session.Stderr = os.Stderr

			tWidth, tHeight, err := terminal.GetSize(fd)
			if err != nil {
				fmt.Println("获取SSH终端窗口大小错误：", err)
				return
			}

			modes := ssh.TerminalModes{
				ssh.ECHO:          1,
				ssh.TTY_OP_ISPEED: 14400,
				ssh.TTY_OP_OSPEED: 14400,
			}

			if err := session.RequestPty("xterm-256color", tHeight, tWidth, modes); err != nil {
				fmt.Println("请求SSH终端窗口错误：", err)
				return
			}

			if err := session.Shell(); err != nil {
				fmt.Println("启动SSH终端窗口错误：", err)
				return
			}

			c.Stop()

			_ = session.Wait()
		},
	}
}
