package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"iOSBox/pkg/idevice"

	"github.com/gookit/gcli/v3"
	"github.com/gookit/gcli/v3/progress"
	"golang.org/x/xerrors"
)

var AppListCommand = &gcli.Command{
	Name:    "apps",
	Desc:    "显示当前设备应用列表",
	Aliases: []string{"as"},
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "应用名称")
	},
	Func: func(c *gcli.Command, args []string) error {
		device, err := idevice.GetDevice()
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}
		conn, err := idevice.NewAppManagerService(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误：%w", err)
		}
		defer conn.Close()

		appList, err := conn.GetApplications()
		if err != nil {
			return err
		}

		c.Println("--------------------------------------------------------------")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 0, 1, ' ', 0)
		for i, info := range appList {
			if len(args) == 1 && args[0] != info.CFBundleDisplayName {
				continue
			}
			_, _ = fmt.Fprintln(w, fmt.Sprintf("Number\t: %d", i))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("Name\t: %s", info.CFBundleDisplayName))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("BundleId\t: %s", info.CFBundleIdentifier))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("Version\t: %s", info.CFBundleShortVersionString))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("Executable\t: %s", info.CFBundleExecutable))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("Container\t: %s", info.Container))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("Path\t: %s", info.Path))
			_, _ = fmt.Fprintln(w, "--------------------------------------------------------------")
		}
		_ = w.Flush()

		return nil
	},
}

var AppInstallCommand = &gcli.Command{
	Name:     "install",
	Aliases:  []string{"ins", "i"},
	Desc:     "安装应用",
	Examples: "{$binName} {$cmd} $HOME/Downloads/example.ipa",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "IPA文件路径", true)
	},
	Func: func(c *gcli.Command, args []string) error {
		if len(args) == 0 {
			return xerrors.Errorf("未传入IPA文件路径")
		}

		device, err := idevice.GetDevice()
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}
		fservice, err := idevice.NewFileManagerService(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误：%w", err)
		}
		defer fservice.Close()

		ipaPath := args[0]
		remotePath := "PublicStaging/" + filepath.Base(ipaPath)
		lfile, err := os.Open(ipaPath)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(lfile)

		fi, _ := lfile.Stat()
		total := fi.Size() / int64(idevice.DefaultChunkSize)

		cs := progress.BarStyles[3]
		p := progress.CustomBar(40, cs)
		p.MaxSteps = uint(total)
		// p.Format = progress.FullBarFormat
		p.AddMessage("正在上传...", "")
		p.Start()
		if err := fservice.FileUpload(lfile, remotePath, func(count int) {
			p.Advance()
		}); err != nil {
			return xerrors.Errorf("IPA文件上传错误：%w", err)
		}
		p.Finish()

		aservice, err := idevice.NewAppManagerService(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误: %w", err)
		}
		defer aservice.Close()

		p = progress.CustomBar(40, cs)
		p.MaxSteps = uint(100)
		p.Format = progress.FullBarFormat
		p.Start()
		if err := aservice.Install(remotePath, func(ret idevice.AppInstallResponse) {
			p.AdvanceTo(uint(ret.PercentComplete))
		}); err != nil {
			return xerrors.Errorf("安装应用错误：%w", err)
		}
		p.Finish()

		return nil
	},
}

var AppUninstallCommand = &gcli.Command{
	Name:     "uninstall",
	Desc:     "卸载应用",
	Aliases:  []string{"uns", "u"},
	Examples: "{$binName} {$cmd} con.xxx.xxx",
	Config: func(c *gcli.Command) {
		c.AddArg("arg0", "应用BundleID", true)
	},
	Func: func(c *gcli.Command, args []string) error {
		if len(args) == 0 {
			return xerrors.Errorf("未传入应用BundleID")
		}

		device, err := idevice.GetDevice()
		if err != nil {
			return xerrors.Errorf("连接iOS设备错误: %w", err)
		}

		aservice, err := idevice.NewAppManagerService(device)
		if err != nil {
			return xerrors.Errorf("连接服务错误: %w", err)
		}
		defer aservice.Close()

		if err := aservice.Uninstall(args[0]); err != nil {
			return xerrors.Errorf("卸载应用[%s]错误：%w", args[0], err)
		}

		c.Println("应用卸载完成")

		return nil
	},
}
