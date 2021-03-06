package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofmt/iOSBox/pkg/idevice"

	"github.com/gookit/gcli/v3"
	"github.com/gookit/gcli/v3/progress"
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

		device, err := idevice.GetDevice()
		if err != nil {
			return err
		}

		service := idevice.NewForwardService(device)
		err = service.Start(uint16(localPort), uint16(22), func(s string, err error) {
			if err != nil {
				fmt.Println(err)
			}
		})
		if err != nil {
			return err
		}
		defer service.Close()

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
		result, err := shellRun(strings.Join(args, " "))
		if err != nil {
			return xerrors.Errorf("执行SHELL命令错误：%w", err)
		}

		c.Println(string(result))

		return nil
	},
}

var LdrestartCommand = &gcli.Command{
	Name: "redaemon",
	Desc: "重启守护进程",
	Func: func(c *gcli.Command, args []string) error {
		if _, err := shellRun("/usr/bin/ldrestart"); err != nil {
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
	Examples: "{$binName} {$cmd} testdata/example.js :/tmp/example.js",
	Func: func(c *gcli.Command, args []string) error {
		return scp(args)
	},
}

func scp(args []string) error {
	// scp remote:./testfile ./testfile
	// scp ./testfile remote:./testfile

	var (
		remotePath   = ""
		localPath    = ""
		copyToRemote = false
	)
	if strings.HasPrefix(args[0], ":") {
		remotePath = strings.TrimLeft(args[0], ":")
		localPath = args[1]
		copyToRemote = false
	} else if strings.HasPrefix(args[1], ":") {
		remotePath = strings.TrimLeft(args[1], ":")
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

	device, err := idevice.GetDevice()
	if err != nil {
		return err
	}

	service := idevice.NewForwardService(device)
	err = service.Start(uint16(localPort), uint16(22), func(s string, err error) {
		if err != nil {
			fmt.Println(err)
		}
	})
	if err != nil {
		return err
	}
	defer service.Close()

	deviceIp := fmt.Sprintf("127.0.0.1:%d", localPort)
	cli, err := newSSHClient(deviceIp)
	if err != nil {
		return xerrors.Errorf("连接SSH错误：%w", err)
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

	if copyToRemote {
		if err := scpTo(session, localPath, remotePath); err != nil {
			if err.Error() == "Process exited with status 1" {
				return nil
			}

			return xerrors.Errorf("SCP错误：%w", err)
		}
	} else {
		if err := scpFrom(session, remotePath, localPath); err != nil {
			return xerrors.Errorf("SCP错误：%w", err)
		}
	}

	return nil
}

func scpTo(session *ssh.Session, local, remote string) error {
	done := make(chan bool)
	cherr := make(chan error)
	go func() {
		err := session.Run("/usr/bin/scp -qrt " + remote)
		if err != nil {
			cherr <- err
			return
		}
	}()

	go func() {
		f, err := os.Open(local)
		if err != nil {
			cherr <- err
			return
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		fi, _ := f.Stat()

		w, err := session.StdinPipe()
		if err != nil {
			cherr <- err
			return
		}
		defer func(w io.WriteCloser) {
			_ = w.Close()
		}(w)

		_, err = fmt.Fprintf(w, "C%04o %d %s\n", fi.Mode(), fi.Size(), filepath.Base(remote))
		if err != nil {
			cherr <- err
			return
		}

		cs := progress.BarStyles[3]
		p := progress.CustomBar(40, cs)
		p.Format = progress.FullBarFormat
		p.MaxSteps = uint(fi.Size() / 4096)
		p.Start()

		buf := make([]byte, 4096)
		for {
			switch nr, err := f.Read(buf[:]); true {
			case nr < 0:
				cherr <- xerrors.Errorf("cat: error reading: %w", err)
				return
			case nr == 0:
				_, err = w.Write([]byte{0})
				if err != nil {
					cherr <- err
					return
				}

				done <- true
				p.Finish()
				return
			case nr > 0:
				if _, err := w.Write(buf); err != nil {
					cherr <- xerrors.Errorf("file write error: %w", err)
					return
				}
				p.Advance()
			}
		}
	}()

	select {
	case err := <-cherr:
		if err != nil {
			return err
		}
	case <-done:
		return nil
	}

	return nil
}

func scpFrom(session *ssh.Session, remote, local string) error {
	done := make(chan bool)
	cherr := make(chan error)
	go func() {
		err := session.Run("/usr/bin/scp -qrf " + remote)
		if err != nil {
			cherr <- err
			return
		}
		done <- true
	}()

	go func() {
		writer, err := session.StdinPipe()
		if err != nil {
			cherr <- err
			return
		}
		defer func(writer io.WriteCloser) {
			_ = writer.Close()
		}(writer)

		reader, err := session.StdoutPipe()
		if err != nil {
			cherr <- err
			return
		}

		_, err = writer.Write([]byte{0})
		if err != nil {
			cherr <- err
			return
		}

		var (
			permMode int
			fileSize int
			fileName string
		)
		_, err = fmt.Fscanf(reader, "C%04o %d %s", &permMode, &fileSize, &fileName)
		if err != nil {
			cherr <- err
			return
		}

		_, err = writer.Write([]byte{0})
		if err != nil {
			cherr <- err
			return
		}

		f, err := os.OpenFile(local, os.O_CREATE|os.O_RDWR, os.FileMode(permMode))
		if err != nil {
			cherr <- err
			return
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		cs := progress.BarStyles[3]
		p := progress.CustomBar(40, cs)
		p.Format = progress.FullBarFormat

		p.Start()

		chunkSize := int64(4096)
		totalSize := int64(fileSize)
		p.MaxSteps = uint(totalSize / chunkSize)
		curSize := int64(0)
		for curSize < totalSize {
			if chunkSize > totalSize-curSize {
				chunkSize = totalSize - curSize
			}
			bs := make([]byte, chunkSize)
			n, err := reader.Read(bs)
			if err != nil {
				cherr <- err
				return
			}
			curSize += int64(n)
			_, err = f.Write(bs[:n])
			if err != nil {
				cherr <- err
				return
			}
			p.Advance()
		}
		p.Finish()

		err = f.Close()
		if err != nil {
			cherr <- err
			return
		}

		_, err = writer.Write([]byte{0})
		if err != nil {
			cherr <- err
			return
		}

	}()

	select {
	case err := <-cherr:
		if err != nil {
			return err
		}
	case <-done:
		return nil
	}

	return nil
}

func shellRun(cmd string) (result []byte, err error) {
	localPort, err := GetAvailablePort()
	if err != nil {
		return
	}

	device, err := idevice.GetDevice()
	if err != nil {
		return
	}

	service := idevice.NewForwardService(device)
	err = service.Start(uint16(localPort), uint16(22), func(s string, err error) {
		if err != nil {
			fmt.Println(err)
		}
	})
	if err != nil {
		return
	}
	defer service.Close()

	deviceIp := fmt.Sprintf("127.0.0.1:%d", localPort)

	cli, err := newSSHClient(deviceIp)
	if err != nil {
		return
	}
	defer func(cli *ssh.Client) {
		_ = cli.Close()
	}(cli)

	session, err := cli.NewSession()
	if err != nil {
		return
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	buf := new(bytes.Buffer)
	session.Stdout = buf
	session.Stderr = os.Stderr
	if err := session.Run(cmd); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
