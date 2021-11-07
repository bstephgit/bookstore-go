package utils

import (
	"os"

	"github.com/rthornton128/goncurses"
)

func init() {
}

func NewTerminal() *Terminal {
	term_name, exist := os.LookupEnv("TERM")
	if !exist || len(term_name) == 0 {
		panic("TERM env not found")
	}

	goncurses.Init()

	screen, err := goncurses.NewTerm(term_name, os.Stdout, os.Stdin)
	if err != nil {
		panic(err)
	}

	goncurses.StdScr().Keypad(true)
	goncurses.StdScr().ScrollOk(true)
	goncurses.CBreak(true)

	term := Terminal{0, 0, screen, []*ScreenContext{}}

	term.Lines, term.Cols = GetScreenSize()
	return &term

}

type ScreenContext struct {
	CurrentScreen Screen
	Scroll        int
	LineSize      int
	ColSize       int
	LogicLines    int
	LogicCols     int
	Reading       bool
	CursorLine    int
	CursorCol     int
}

type Terminal struct {
	Cols        int
	Lines       int
	Screen      *goncurses.Screen
	ScreenStack []*ScreenContext
}

func (t *Terminal) ClearScreen() {
	goncurses.StdScr().Clear()
	goncurses.StdScr().Refresh()
}

func GetScreenSize() (int, int) {
	y, x := goncurses.StdScr().MaxYX()

	return y, x
}
func (t *Terminal) MoveCursorTo(ypos int, xpos int) {
	goncurses.StdScr().Move(ypos, xpos)
	goncurses.StdScr().Refresh()
}

func (t *Terminal) PrintMessage(s string, p ...interface{}) {
	y, x := goncurses.StdScr().CursorYX()
	goncurses.StdScr().MovePrintf(y+1, 0, s, p...)
	goncurses.StdScr().Move(y, x)
	goncurses.StdScr().Refresh()
}

func (t *Terminal) Printf(s string, p ...interface{}) {
	goncurses.StdScr().Printf(s, p...)
	goncurses.StdScr().Refresh()
}

func (t *Terminal) Println(s string) {
	goncurses.StdScr().Println(s)
	goncurses.StdScr().Refresh()
}

func (t *Terminal) GetChar() goncurses.Key {
	k := goncurses.StdScr().GetChar()
	return k
}

func (t *Terminal) ScrollScr(n int) {

	ctx := t.ScreenStack[len(t.ScreenStack)-1]

	if ctx.LogicLines > t.Lines {
		scroll := ctx.Scroll
		ctx.Scroll += n

		if ctx.Scroll < 0 {
			ctx.Scroll = 0
		}

		if (ctx.Scroll + t.Lines) > ctx.LogicLines-1 {
			ctx.Scroll = ctx.LogicLines - t.Lines
		}

		n = ctx.Scroll - scroll
		goncurses.StdScr().Scroll(n)
		t.CurrentContext().CurrentScreen.OnScroll(n)
	}
}

func (t *Terminal) NewScreen(scr Screen) {

	if t.CurrentContext() != nil {
		t.SaveCursorPos()
	}
	ctx := &ScreenContext{scr, 0, 1, 1, t.Lines, t.Cols, false, 0, 0}

	// add new context
	t.ScreenStack = append(t.ScreenStack, ctx)
	// init screen
	scr.Init(t, ctx)
	goncurses.StdScr().Resize(ctx.LogicLines*ctx.LineSize, t.Cols*ctx.ColSize)
	// run the screen
	scr.Run()
	t.ClearScreen()

	// screen is over, remove context
	t.ScreenStack = t.ScreenStack[0 : len(t.ScreenStack)-1]
	// restore previous context
	ctx = t.CurrentContext()
	if ctx != nil {
		goncurses.StdScr().Resize(ctx.LogicLines*ctx.LineSize, t.Cols*ctx.ColSize)
		ctx.CurrentScreen.OnRefresh(0, 0)
		t.ScrollScr(ctx.Scroll)
		t.MoveCursorTo(ctx.CursorLine, ctx.CursorCol)
	}
}

func (tty *Terminal) CurrentContext() *ScreenContext {
	if len(tty.ScreenStack) > 0 {
		return tty.ScreenStack[len(tty.ScreenStack)-1]
	}
	return nil
}

func (t *Terminal) BeginRead() {
	ctx := t.CurrentContext()
	if ctx.Reading {
		panic("Cannot start reading: already reading")
	}
	ctx.Reading = true
	for ctx.Reading {
		ctx.CurrentScreen.OnKey(t.GetChar())
	}
}

func (t *Terminal) EndRead() {
	ctx := t.CurrentContext()
	ctx.Reading = false
}

func (t *Terminal) MoveNextCol() {
	ctx := t.CurrentContext()
	if ctx.CursorCol+1 < ctx.LogicCols {
		ctx.CursorCol += 1
		t.MoveCursorTo(ctx.CursorLine*ctx.LineSize, ctx.CursorCol*ctx.ColSize)
	}
}

func (t *Terminal) MovePrevCol() {

	ctx := t.CurrentContext()
	ctx.CursorCol -= 1
	if ctx.CursorCol < 0 {
		ctx.CursorCol = 0
	} else {
		t.MoveCursorTo(ctx.CursorLine*ctx.LineSize, ctx.CursorCol*ctx.ColSize)
	}
}

func (t *Terminal) MoveNextLine() {

	ctx := t.CurrentContext()

	next_line_scr := (ctx.CursorLine + 1) * ctx.LineSize

	// if cursor not at bottom of screen
	if next_line_scr < t.Lines {
		ctx.CursorLine += 1
		t.MoveCursorTo(ctx.CursorLine*ctx.LineSize, ctx.CursorCol*ctx.ColSize)
	} else if ctx.CursorLine+ctx.Scroll+1 < ctx.LogicLines {
		ctx.Scroll += 1
		goncurses.StdScr().Scroll(1)
		ctx.CurrentScreen.OnScroll(1)
	}

}

func (t *Terminal) MovePrevLine() {

	ctx := t.CurrentContext()

	// if cursor not at top of screen
	if ctx.CursorLine-1 > -1 {
		ctx.CursorLine -= 1
		t.MoveCursorTo(ctx.CursorLine*ctx.LineSize, ctx.CursorCol*ctx.ColSize)
	} else if ctx.Scroll-1 > -1 {
		ctx.Scroll -= 1
		goncurses.StdScr().Scroll(-1)
		ctx.CurrentScreen.OnScroll(-1)
	}
}

func (t *Terminal) CursorAddress(line, col int) {

	ctx := t.CurrentContext()

	if line < 0 {
		line = 0
	}
	if line > ctx.LogicLines-1 {
		line = ctx.LogicLines - 1
	}

	if col < 0 {
		col = 0
	}
	if col > ctx.LogicCols-1 {
		col = ctx.LogicCols - 1
	}

	t.MoveCursorTo(line*ctx.LineSize, col*ctx.ColSize)
}

func (t *Terminal) SaveCursorPos() (int, int) {
	y, x := goncurses.StdScr().CursorYX()
	ctx := t.CurrentContext()
	ctx.CursorLine = y / ctx.LineSize
	ctx.CursorCol = x / ctx.ColSize
	return y, x
}

func TerminalLoop() {

	tty := NewTerminal()
	tty.NewScreen(&MenuScreen{})
	tty.Screen.End()
}
