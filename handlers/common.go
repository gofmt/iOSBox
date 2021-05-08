package handlers

import (
	"github.com/gookit/gcli/v3"
)

func init() {
	gcli.AppHelpTemplate = `{{.Desc}} (版本: <info>{{.Version}}</>)
-----------------------------------------------------{{range $cmdName, $c := .Cs}}
  <info>{{$c.Name | paddingName }}</> {{$c.HelpDesc}}{{if $c.Aliases}} (别名: <green>{{ join $c.Aliases ","}}</>){{end}}{{end}}
  <info>{{ paddingName "help" }}</> 显示帮助信息

使用 "<cyan>{$binName} COMMAND -h</>" 查看命令的其他帮助信息
`

	gcli.CmdHelpTemplate = `{{.Desc}}
	
<comment>用法:</>
  {$binName} [global options] {{if .Cmd.NotStandalone}}<cyan>{{.Cmd.Path}}</> {{end}}[--options ...] [arguments ...]{{ if .Subs }}
  {$binName} [global options] {{if .Cmd.NotStandalone}}<cyan>{{.Cmd.Path}}</> {{end}}<cyan>SUBCOMMAND</> [--options ...] [arguments ...]{{end}}
{{if .Options}}
<comment>Options:</>
{{.Options}}{{end}}{{if .Cmd.Args}}
<comment>Arguments:</>{{range $a := .Cmd.Args}}
  <info>{{$a.HelpName | printf "%-12s"}}</>{{$a.Desc | ucFirst}}{{if $a.Required}}<red>*</>{{end}}{{end}}
{{end}}{{ if .Subs }}
<comment>命令列表:</>{{range $n,$c := .Subs}}
  <info>{{$c.Name | paddingName }}</> {{$c.HelpDesc}}{{if $c.Aliases}} (alias: <green>{{ join $c.Aliases ","}}</>){{end}}{{end}}
{{end}}{{if .Cmd.Examples}}
<comment>Examples:</>
{{.Cmd.Examples}}{{end}}{{if .Cmd.Help}}
<comment>Help:</>
{{.Cmd.Help}}{{end}}`
}

func helpRender(c *gcli.Command) {

}
