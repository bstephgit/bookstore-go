package console

import (
	"strings"

	"github.com/bookstore-go/download"
	"github.com/bookstore-go/utils"
	"github.com/rthornton128/goncurses"
)

type Screen interface {
	Init(tty *Terminal, sc *ScreenContext)
	Run()
	OnScroll(y int)
	OnKey(k goncurses.Key)
	OnRefresh(Lines, Cols int)
}

type MenuScreen struct {
	Tty              *Terminal
	CursorX, CursorY int
}

func (menu *MenuScreen) PrintMenu() {

	Title := `
    __________               __      _________ __                        
    \______   \ ____   ____ |  | __ /   _____//  |_  ___________   ____  
     |    |  _//  _ \ /  _ \|  |/ / \_____  \\   __\/  _ \_  __ \_/ __ \ 
     |    |   (  <_> |  <_> )    <  /        \|  | (  <_> )  | \/\  ___/ 
     |______  /\____/ \____/|__|_ \/_______  /|__|  \____/|__|    \___  >
            \/                   \/        \/                         \/ 
	`

	line1 := "         ------------------ Menu ------------------"
	line2 := "                  a - Display subjects"
	line3 := "                  b - Search books by subjects"
	line4 := "                  c - Get book info"
	line5 := "                  d - Download book"
	line6 := "                  q - Quit"

	menu.Tty.Println(Title)
	menu.Tty.Println(line1)
	menu.Tty.Println("")
	menu.Tty.Println(line2)
	menu.Tty.Println(line3)
	menu.Tty.Println(line4)
	menu.Tty.Println(line5)
	menu.Tty.Println(line6)
	menu.Tty.Printf("> ")
}

func (menu *MenuScreen) Init(tty *Terminal, ctx *ScreenContext) {
	menu.Tty = tty
}

func (menu *MenuScreen) Run() {
	menu.Tty.ClearScreen()
	menu.PrintMenu()
	menu.CursorY, menu.CursorX = menu.Tty.SaveCursorPos()
	menu.Tty.BeginRead()
}

func (menu *MenuScreen) OnScroll(y int) {
	// no scroll
}

func (menu *MenuScreen) OnKey(key goncurses.Key) {
	if key == 'a' {
		menu.Tty.NewScreen(&SubjectsScreen{})
	} else if key == 'q' {
		menu.Tty.PrintMessage("Quit application...")
		menu.Tty.EndRead()
	} else if key == 'b' {
		menu.Tty.PrintMessage("Search books by subjects")
	} else if key == 'c' {
		menu.Tty.PrintMessage("Get book info")
	} else if key == 'd' {
		menu.Tty.PrintMessage("Download book")

	} else {
		menu.Tty.PrintMessage("Unrecognized command %d\n", key)
	}
	menu.Tty.CursorAddress(menu.CursorY, menu.CursorX)
}

func (menu *MenuScreen) OnRefresh(Lines, Cols int) {
	menu.Tty.ClearScreen()
	menu.PrintMenu()
}

type SubjectsScreen struct {
	Subjects []utils.Subject
	Tty      *Terminal
}

func (subscr *SubjectsScreen) PrintSubjects() {
	ctx := subscr.Tty.CurrentContext()
	for i := range subscr.Subjects {

		line := i % ctx.LogicLines
		col := i / ctx.LogicLines
		subscr.Tty.CursorAddress(line, col)

		subscr.PrintItem(i)
		//tty.Printf("%s (%d)", subscr.Subjects[i].Name, subscr.Subjects[i].Id)
		//tty.Printf("(%d[%d,%d])", i, line, col)
	}
}

func (subscr *SubjectsScreen) PrintItem(index int) {
	subscr.Tty.Printf("%s (%d)", subscr.Subjects[index].Name, subscr.Subjects[index].Id)
}

func (subscr *SubjectsScreen) Init(tty *Terminal, ctx *ScreenContext) {

	subjects, err := utils.GetSubjects()
	if err != nil {
		tty.Printf("%v", err)
	}
	ctx.ColSize = 35

	if tty.Cols < ctx.ColSize {
		ctx.ColSize = tty.Cols
	}
	ctx.LogicCols = tty.Cols / ctx.ColSize
	ctx.LogicLines = len(subjects) / ctx.LogicCols

	if (ctx.LogicCols * ctx.LogicLines) < len(subjects) {
		ctx.LogicLines += 1
	}

	tty.ClearScreen()

	subscr.Tty = tty
	subscr.Subjects = subjects
}

func (subscr *SubjectsScreen) Run() {

	subscr.OnRefresh(0, 0)
	if len(subscr.Subjects) > 0 {
		subscr.Tty.Highlight(0, 0, subscr.Subjects[0].Name, true)
	}
	subscr.Tty.BeginRead()
}

func (subscr *SubjectsScreen) OnScroll(y int) {

	ctx := subscr.Tty.CurrentContext()
	if y < 0 {
		y = -y

		line := subscr.Tty.CurrentContext().CursorLine
		col := subscr.Tty.CurrentContext().CursorCol

		for i := 0; i < y; i += 1 {
			for j := 0; j < ctx.LogicCols; j += 1 {
				subscr.Tty.CursorAddress(i, j)
				index := (ctx.LogicLines * j) + ctx.Scroll + i
				subscr.PrintItem(index)
			}
		}
		subscr.Tty.CursorAddress(line, col)

	}
}

func (subscr *SubjectsScreen) OnKey(key goncurses.Key) {
	ctx := subscr.Tty.CurrentContext()
	index := (ctx.CursorCol * ctx.LogicLines) + ctx.CursorLine + ctx.Scroll
	if index < len(subscr.Subjects) {
		subscr.Tty.Highlight(ctx.CursorLine, ctx.CursorCol, subscr.Subjects[index].Name, false)
	}

	switch key {
	case 'q':
		subscr.Tty.EndRead()
	case goncurses.KEY_ESC:
		subscr.Tty.EndRead()
	case goncurses.KEY_UP:
		subscr.Tty.MovePrevLine()
	case goncurses.KEY_DOWN:
		subscr.Tty.MoveNextLine()
	case goncurses.KEY_RIGHT:
		subscr.Tty.MoveNextCol()
	case goncurses.KEY_LEFT:
		subscr.Tty.MovePrevCol()
	case goncurses.KEY_RETURN:
		ctx := subscr.Tty.CurrentContext()
		index := (ctx.CursorCol * ctx.LogicLines) + ctx.CursorLine + ctx.Scroll
		if index < len(subscr.Subjects) {
			sb := &SubjectBooks{}
			sb.Sub = &subscr.Subjects[index]
			subscr.Tty.NewScreen(sb)
		}
	}
	index = (ctx.CursorCol * ctx.LogicLines) + ctx.CursorLine + ctx.Scroll
	if index < len(subscr.Subjects) {
		subscr.Tty.Highlight(ctx.CursorLine, ctx.CursorCol, subscr.Subjects[index].Name, true)
	}
}

func (subscr *SubjectsScreen) OnRefresh(Lines, Cols int) {
	subscr.Tty.ClearScreen()
	subscr.PrintSubjects()
	subscr.Tty.CursorAddress(0, 0)
}

type SubjectBooks struct {
	Tty       *Terminal
	BookLines []utils.BookLine
	Sub       *utils.Subject
}

func (subb *SubjectBooks) Init(t *Terminal, ctx *ScreenContext) {

	subb.BookLines, _ = utils.GetSubjectBooks(int(subb.Sub.Id))
	subb.Tty = t

	ctx.LogicLines = len(subb.BookLines)

	subb.Tty.ClearScreen()
}

func (subb *SubjectBooks) Run() {
	subb.Tty.CursorAddress(0, 0)
	subb.OnRefresh(0, 0)
	if len(subb.BookLines) > 0 {
		subb.Tty.Highlight(0, 0, subb.BookLines[0].Title, true)
	}
	subb.Tty.BeginRead()
}

func (subb *SubjectBooks) OnRefresh(lines, cols int) {
	for line, book := range subb.BookLines {
		subb.Tty.CursorAddress(line, 0)
		subb.Tty.Printf("%s", book.Title)
	}
	subb.Tty.CursorAddress(0, 0)
}

func (subb *SubjectBooks) OnScroll(y int) {
	if y < 0 {
		y = -y
		line := subb.Tty.CurrentContext().CursorLine
		subb.Tty.CursorAddress(0, 0)
		for index := 0; index < y; index += 1 {
			subb.Tty.CursorAddress(index, 0)
			book_index := index + subb.Tty.CurrentContext().Scroll
			subb.Tty.Printf("%s", subb.BookLines[book_index].Title)
		}
		subb.Tty.CursorAddress(line, 0)
	}
}

func (subb *SubjectBooks) OnKey(k goncurses.Key) {
	ctx := subb.Tty.CurrentContext()
	index := ctx.CursorLine + ctx.Scroll
	subb.Tty.Highlight(ctx.CursorLine, 0, subb.BookLines[index].Title, false)

	switch k {
	case goncurses.KEY_ESC:
	case 'q':
		subb.Tty.EndRead()
	case goncurses.KEY_DOWN:
		subb.Tty.MoveNextLine()
	case goncurses.KEY_UP:
		subb.Tty.MovePrevLine()
	case goncurses.KEY_PAGEDOWN:
		subb.Tty.ScrollScr(subb.Tty.Lines)
	case goncurses.KEY_PAGEUP:
		subb.Tty.ScrollScr(-subb.Tty.Lines)
	case goncurses.KEY_RETURN:
		line := subb.Tty.CurrentContext().CursorLine + subb.Tty.CurrentContext().Scroll
		if line >= 0 && line < len(subb.BookLines) {
			bs := &BookScreen{subb.Tty, int(subb.BookLines[line].Id), nil, nil, ""}
			subb.Tty.NewScreen(bs)
		}
	}
	index = ctx.CursorLine + ctx.Scroll
	subb.Tty.Highlight(ctx.CursorLine, 0, subb.BookLines[index].Title, true)
}

type BookScreen struct {
	Tty     *Terminal
	BookId  int
	BookObj *utils.Book
	Err     error
	Text    string
}

func (bookscr *BookScreen) Init(t *Terminal, ctx *ScreenContext) {
	bookscr.Tty = t
	bookscr.BookObj, bookscr.Err = utils.GetBook(bookscr.BookId)

	bookscr.Text = t.FormatText(bookscr.BookObj.Description)

	ctx.LogicLines = strings.Count(bookscr.Text, "\n") + 8
	t.ClearScreen()
}

func (bookscr *BookScreen) Run() {
	bookscr.OnRefresh(0, 0)
	bookscr.Tty.BeginRead()
}

func (bookscr *BookScreen) OnScroll(y int) {

	if y < 0 {
		y = -y

		line := bookscr.Tty.CurrentContext().CursorLine
		scroll := bookscr.Tty.CurrentContext().Scroll
		bookscr.Tty.CursorAddress(0, 0)

		for line := 0; line < y; line += 1 {

			if line+scroll > 7 {

			} else {

			}
		}

		bookscr.Tty.CursorAddress(line, 0)
	}

}

func (tty *Terminal) PrintTitle(title string) {
	tty.CursorAddress(1, 0)
	tty.Printf("Title:")
	tty.CursorAddress(1, 30)
	tty.Printf("\"%s\"", title)
}

func (tty *Terminal) PrintAuthors(authors string) {
	tty.CursorAddress(2, 0)
	tty.Printf("Authors:")
	tty.CursorAddress(2, 30)
	tty.Printf("\"%s\"", authors)
}

func (tty *Terminal) PrintYear(year int) {
	tty.CursorAddress(3, 0)
	tty.Printf("Publication year:")
	tty.CursorAddress(3, 30)
	tty.Printf("%d", year)
}

func (tty *Terminal) PrintDescription(text string) {
	tty.CursorAddress(5, 0)
	tty.Printf("Description: ")
	tty.CursorAddress(7, 0)
	tty.Printf("%s\n(%d lines)", text, tty.CurrentContext().LogicLines)
}

func (tty *Terminal) FormatText(text string) string {
	re := strings.NewReplacer("<BR>", "\n", "<br>", "\n", "<p>", "\n", "</p>", "", "<ul>", "\n", "</ul>", "\n", "<li>", "\t+ ", "</li>", "\n")
	formatted := re.Replace(text)
	var curword, curline, all strings.Builder

	if false {
		return formatted
	}
	for _, r := range formatted {

		switch r {
		case ' ':
			if curline.Len()+curword.Len() >= tty.Cols-1 {
				curline.WriteString("\n")
				all.WriteString(curline.String())
				curline.Reset()
			}
			curline.WriteString(curword.String())
			curword.Reset()
			if curline.Len()+1 < tty.Cols {
				curline.WriteRune(r)
			}
		case '\n':
			if curline.Len()+1 < tty.Cols {
				curline.WriteString(" ")
			} else {
				curline.WriteString("\n")
				all.WriteString(curline.String())
				curline.Reset()
			}

		case '\t':
			// handle word
			if curline.Len()+curword.Len() >= tty.Cols-1 {
				all.WriteString("\n")
				all.WriteString(curline.String())
				curline.Reset()
			}
			curline.WriteString(curword.String())
			curword.Reset()

			tabstr := "    "
			if curline.Len()+len(tabstr) < tty.Cols-1 {
				curline.WriteString(tabstr)
			} else {
				all.WriteString("\n")
				all.WriteString(curline.String())
				curline.Reset()
			}

		default:
			curword.WriteRune(r)
		}
	}

	return all.String()
}

func (bookscr *BookScreen) OnRefresh(lins, cols int) {
	tty := bookscr.Tty

	tty.ClearScreen()

	book := bookscr.BookObj

	if book != nil {
		tty.PrintTitle(book.Title)
		tty.PrintAuthors(book.Authors)
		tty.PrintYear(book.Year)
		tty.PrintDescription(bookscr.Text)
	}
	if bookscr.Err != nil {
		tty.Printf("%v", bookscr.Err)
	}
	tty.CursorAddress(0, 0)
}

func (bookscr *BookScreen) OnKey(k goncurses.Key) {
	switch k {
	case 'q':
		bookscr.Tty.EndRead()
	case 'd':
		dl := &DownloadScreen{}
		dl.BookId = bookscr.BookId
		bookscr.Tty.NewScreen(dl)
	case goncurses.KEY_PAGEDOWN:
		bookscr.Tty.ScrollScr(bookscr.Tty.Lines)
	case goncurses.KEY_PAGEUP:
		bookscr.Tty.ScrollScr(-bookscr.Tty.Lines)
	}
}

type DownloadScreen struct {
	Tty    *Terminal
	BookId int
	BookDl *utils.BookDownload
	Done   bool
}

func (ds *DownloadScreen) Init(tty *Terminal, ctx *ScreenContext) {
	ds.Tty = tty
	ds.BookDl, _ = utils.GetDownloadInfo(ds.BookId)
	tty.ClearScreen()
	ds.Done = false
}

func (ds *DownloadScreen) Run() {
	ds.OnRefresh(0, 0)
	ds.Tty.BeginRead()
}

func (ds *DownloadScreen) OnKey(key goncurses.Key) {
	switch key {
	case 'n':
		ds.Tty.EndRead()
	case 'y':
		err := download.DownloadFile(ds.BookDl)
		//ds.Tty.EndRead()
		if err != nil {
			ds.Tty.CursorAddress(7, 0)
			ds.Tty.Printf("%v\n", err)
		}
		ds.Done = true
		ds.Tty.Println("Press any key to return to book page")
	default:
		if ds.Done {
			ds.Tty.EndRead()
		}
	}

}

func (ds *DownloadScreen) OnScroll(y int) {
	// nop
}

func (ds *DownloadScreen) OnRefresh(lines, cols int) {
	ds.Tty.CursorAddress(0, 0)
	ds.Tty.Printf("File:")
	ds.Tty.CursorAddress(0, 30)
	ds.Tty.Printf("%s", ds.BookDl.FileName)
	ds.Tty.CursorAddress(1, 0)
	ds.Tty.Printf("Size:")
	ds.Tty.CursorAddress(1, 30)
	ds.Tty.Printf("%d", ds.BookDl.FileSize)
	ds.Tty.CursorAddress(2, 0)
	ds.Tty.Printf("Storage vendor:")
	ds.Tty.CursorAddress(2, 30)
	ds.Tty.Printf("%s (%s)", ds.BookDl.Vendor, ds.BookDl.VendorCode)

	ds.Tty.CursorAddress(4, 0)
	ds.Tty.Println("Download book? Y/n")
}
