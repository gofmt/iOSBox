package lldb

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

var (
	app    = &views.Application{}
	window = &mainWindow{}
)

type mainWindow struct {
	views.BoxLayout
}

func (m *mainWindow) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyCtrlC:
			app.Quit()
			return true
		}
	}

	return m.BoxLayout.HandleEvent(ev)
}

func (m *mainWindow) Draw() {
	m.BoxLayout.Draw()
}

func Run() error {
	mainBox := views.NewBoxLayout(views.Horizontal)

	l := views.NewText()
	m := views.NewText()
	r := views.NewText()

	l.SetText("Left (0.0)")
	m.SetText("Middle (0.7)")
	r.SetText("Right (0.3)")
	l.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).
		Background(tcell.ColorRed))
	m.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).
		Background(tcell.ColorLime))
	r.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlue))
	l.SetAlignment(views.AlignBegin)
	m.SetAlignment(views.AlignMiddle)
	r.SetAlignment(views.AlignEnd)

	mainBox.AddWidget(l, 0)
	mainBox.AddWidget(m, 0.7)
	mainBox.AddWidget(r, 0.3)

	title := views.NewTextBar()
	title.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorGray).
		Foreground(tcell.ColorWhite))
	title.SetLeft("lldb 浏览器", tcell.StyleDefault)
	title.SetRight("CTRL+C 退出", tcell.StyleDefault)

	window.SetOrientation(views.Vertical)
	window.AddWidget(title, 0)
	window.AddWidget(mainBox, 1)

	app.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	app.SetRootWidget(window)

	return app.Run()
}
