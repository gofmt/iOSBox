package handlers

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/dtylman/scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
)

func CmdShell(deviceInfo *DeviceInfo) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "ssh",
		Help: "创建SSH交互式环境，需要越狱",
		Func: func(c *ishell.Context) {
			cli, err := newSSHClient(deviceInfo.Entry)
			if err != nil {
				fmt.Println("连接SSH错误：", err)
				return
			}
			defer func(cli *ssh.Client) {
				_ = cli.Close()
			}(cli)

			session, err := cli.NewSession()
			if err != nil {
				fmt.Println("获取SSH会话错误：", err)
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

func CmdShellRun(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "run",
		Help: "执行任何shell命令",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				fmt.Println("参数错误")
				return
			}

			cli, err := newSSHClient(entry)
			if err != nil {
				fmt.Println("创建SSH客户端错误：", err)
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

			session.Stdout = os.Stdout

			if err := session.Run(strings.Join(c.Args, " ")); err != nil {
				fmt.Println("执行SHELL命令错误：", err)
			}
		},
	}
}

func CmdSCP(entry ios.DeviceEntry) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "scp",
		Help: "双向专递文件",
		Func: func(c *ishell.Context) {
			// scp remote:./testfile ./testfile
			// scp ./testfile remote:./testfile
			if len(c.Args) < 2 {
				fmt.Println("参数错误")
				return
			}

			var (
				remotePath   = ""
				localPath    = ""
				copyToRemote = false
			)
			if strings.HasPrefix(c.Args[0], "remote:") {
				remotePath = strings.TrimLeft(c.Args[0], "remote:")
				localPath = c.Args[1]
				copyToRemote = false
			} else if strings.HasPrefix(c.Args[1], "remote:") {
				remotePath = strings.TrimLeft(c.Args[1], "remote:")
				localPath = c.Args[0]
				copyToRemote = true
			}

			if len(remotePath) == 0 || len(localPath) == 0 {
				fmt.Println("SCP参数错误")
				return
			}

			cli, err := newSSHClient(entry)
			if err != nil {
				fmt.Println("连接SSH错误：", err)
				return
			}
			defer func(cli *ssh.Client) {
				_ = cli.Close()
			}(cli)

			if copyToRemote {
				if _, err := scp.CopyTo(cli, localPath, remotePath); err != nil {
					fmt.Println("SCP错误：", err)
					return
				}
			} else {
				if _, err := scp.CopyFrom(cli, remotePath, localPath); err != nil {
					fmt.Println("SCP错误：", err)
					return
				}
			}
		},
	}
}

func newSSHClient(entry ios.DeviceEntry) (*ssh.Client, error) {
	if DeviceIP == "" {
		DeviceIP = GetDeviceIPAddress(entry)
		if DeviceIP == "" {
			return nil, xerrors.New("未获取到设备IP地址")
		}
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

	return ssh.Dial("tcp", fmt.Sprintf("%s:22", DeviceIP), &cfg)
}
