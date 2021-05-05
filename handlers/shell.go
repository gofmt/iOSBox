package handlers

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/forward"
	"github.com/dtylman/scp"
	"github.com/gookit/gcli/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
)

var ShellCommand = &gcli.Command{
	Name: "shell",
	Desc: "创建SSH交互环境",
	Func: func(c *gcli.Command, args []string) error {
		localPort, err := GetAvailablePort()
		if err != nil {
			return err
		}

		device, err := ios.GetDevice("")
		if err != nil {
			return err
		}

		err = forward.Forward(device, uint16(localPort), uint16(22))
		if err != nil {
			return err
		}

		deviceIp := fmt.Sprintf("127.0.0.1:%d", localPort)
		cli, err := newSSHClient(deviceIp)
		if err != nil {
			return xerrors.Errorf("创建SSH客户端错误：%w", err)
		}
		defer func(cli *ssh.Client) {
			_ = cli.Close()
		}(cli)

		session, err := cli.NewSession()
		if err != nil {
			return xerrors.Errorf("获取SSH会话错误：%w", err)
		}
		defer func(session *ssh.Session) {
			_ = session.Close()
		}(session)

		fd := int(os.Stdin.Fd())
		oldState, err := terminal.MakeRaw(fd)
		if err != nil {
			return xerrors.Errorf("创建SSH终端错误：%w", err)
		}
		defer func(fd int, oldState *terminal.State) {
			_ = terminal.Restore(fd, oldState)
		}(fd, oldState)

		session.Stdout = os.Stdout
		session.Stdin = os.Stdin
		session.Stderr = os.Stderr

		tWidth, tHeight, err := terminal.GetSize(fd)
		if err != nil {
			return xerrors.Errorf("获取SSH终端窗口大小错误：%w", err)
		}

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}

		if err := session.RequestPty("xterm-256color", tHeight, tWidth, modes); err != nil {
			return xerrors.Errorf("请求SSH终端窗口错误：%w", err)
		}

		if err := session.Shell(); err != nil {
			return xerrors.Errorf("启动SSH终端窗口错误：%w", err)
		}

		_ = session.Wait()

		return nil
	},
}

var RunCommand = &gcli.Command{
	Name: "run",
	Desc: "执行任何SHELL命令",
	Config: func(c *gcli.Command) {
		c.AddArg("arrArg", "SHELL命令列表", true, true)
	},
	Examples: "{$binName} {$cmd} ls -lah",
	Func: func(c *gcli.Command, args []string) error {
		if err := shellRun(strings.Join(args, " ")); err != nil {
			return xerrors.Errorf("执行SHELL命令错误：%w", err)
		}

		return nil
	},
}

var LdrestartCommand = &gcli.Command{
	Name: "redaemon",
	Desc: "重启守护进程",
	Func: func(c *gcli.Command, args []string) error {
		if err := shellRun("/usr/bin/ldrestart"); err != nil {
			return xerrors.Errorf("执行SHELL命令错误：%w", err)
		}

		return nil
	},
}

var SCPCommand = &gcli.Command{
	Name: "scp",
	Desc: "通过SSH传递文件",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "本地或远程文件路径", true)
		c.AddArg("arg1", "本地或远程文件路径", true)
	},
	Examples: "{$binName} {$cmd} testdata/example.js remote:/tmp/example.js",
	Func: func(c *gcli.Command, args []string) error {
		// scp remote:./testfile ./testfile
		// scp ./testfile remote:./testfile

		var (
			remotePath   = ""
			localPath    = ""
			copyToRemote = false
		)
		if strings.HasPrefix(args[0], "remote:") {
			remotePath = strings.TrimLeft(args[0], "remote:")
			localPath = args[1]
			copyToRemote = false
		} else if strings.HasPrefix(args[1], "remote:") {
			remotePath = strings.TrimLeft(args[1], "remote:")
			localPath = args[0]
			copyToRemote = true
		}

		if len(remotePath) == 0 || len(localPath) == 0 {
			return xerrors.New("SCP参数错误")
		}

		localPort, err := GetAvailablePort()
		if err != nil {
			return err
		}

		device, err := ios.GetDevice("")
		if err != nil {
			return err
		}

		err = forward.Forward(device, uint16(localPort), uint16(22))
		if err != nil {
			return err
		}

		deviceIp := fmt.Sprintf("127.0.0.1:%d", localPort)
		cli, err := newSSHClient(deviceIp)
		if err != nil {
			return xerrors.Errorf("连接SSH错误：%w", err)
		}
		defer func(cli *ssh.Client) {
			_ = cli.Close()
		}(cli)

		if copyToRemote {
			if _, err := scp.CopyTo(cli, localPath, remotePath); err != nil {
				return xerrors.Errorf("SCP错误：%w", err)
			}
		} else {
			if _, err := scp.CopyFrom(cli, remotePath, localPath); err != nil {
				return xerrors.Errorf("SCP错误：%w", err)
			}
		}
		return nil
	},
}

func shellRun(cmd string) error {
	localPort, err := GetAvailablePort()
	if err != nil {
		return err
	}

	device, err := ios.GetDevice("")
	if err != nil {
		return err
	}

	err = forward.Forward(device, uint16(localPort), uint16(22))
	if err != nil {
		return err
	}

	deviceIp := fmt.Sprintf("127.0.0.1:%d", localPort)

	cli, err := newSSHClient(deviceIp)
	if err != nil {
		return err
	}
	defer func(cli *ssh.Client) {
		_ = cli.Close()
	}(cli)

	session, err := cli.NewSession()
	if err != nil {
		return err
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	session.Stdout = os.Stdout

	return session.Run(cmd)
}

func newSSHClient(deviceIp string) (*ssh.Client, error) {
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

	return ssh.Dial("tcp", deviceIp, &cfg)
}

func GetAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer func(listener *net.TCPListener) {
		_ = listener.Close()
	}(listener)

	return listener.Addr().(*net.TCPAddr).Port, nil
}
