package main

import (
	"log"
	"math"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var keyMap = map[int16]string{
	13:  "Enter",
	257: "Up",
	258: "Down",
	256: "Rune",
	27:  "Esc",
}

var inputStyle = tcell.StyleDefault.Foreground(tcell.ColorPink).Background(tcell.ColorReset)
var titleStyle = tcell.StyleDefault.Foreground(tcell.ColorReset).Background(tcell.ColorBlue)
var footerStyle = tcell.StyleDefault.Foreground(tcell.ColorReset).Background(tcell.ColorBlue)
var menuItemStyle = tcell.StyleDefault.Foreground(tcell.ColorReset).Background(tcell.ColorReset)
var menuActiveItemStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorPurple)
var receivedMsgStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
var myMsgStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

type UI struct {
	tcs          tcell.Screen
	screen       Screen
	getKeysMap   func() *KeyEventMap
	enableVMenu  bool
	enableTyping bool
	emptyRow     int
	typed        string
	runes        []rune
}

func NewUI(f func() *KeyEventMap) (*UI, error) {
	ui := UI{getKeysMap: f}
	err := ui.init()
	if err != nil {
		return nil, err
	}
	return &ui, nil
}

func (ui *UI) init() error {
	var err error
	defer func() {
		if err != nil {
			ui.Destroy()
		}
	}()

	ui.tcs, err = tcell.NewScreen()
	if err != nil {
		return err
	}
	if err = ui.tcs.Init(); err != nil {
		return err
	}
	ui.tcs.Clear()
	return nil
}

func (ui *UI) SetScreen(s Screen, typing bool, vMenu bool) {
	ui.screen = s
	ui.enableTyping = typing
	ui.enableVMenu = vMenu
	ui.typed = ""
	ui.runes = []rune{}
	ui.emptyRow = 0
}

func (ui *UI) EnableTyping() {
	ui.enableTyping = true
}

func (ui *UI) DisableTyping() {
	ui.enableTyping = false
}

func (ui *UI) ClearTyped() {
	ui.typed = ""
	ui.runes = []rune{}
}

func (ui *UI) EnableVMenu() {
	ui.enableVMenu = true
}

func (ui *UI) DisableVMenu() {
	ui.enableVMenu = false
}

func (ui *UI) Draw() {
	ui.tcs.Clear()
	ui.emptyRow = 0
	ui.screen.Draw()
	ui.tcs.Show()
}

func (ui *UI) Listen(c chan<- *Event) {
	for {
		// Process event
		ev := ui.tcs.PollEvent()
		// log.Printf("%t", ev)
		switch ev := ev.(type) {
		case *tcell.EventResize:
			ui.tcs.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				panic("err")
			}
			log.Print("Input | ", ev.Name(), " | ", ev.Key(), " | ", ev.Rune())
			if ui.enableTyping {
				if ev.Key() == tcell.KeyRune {
					ui.typed += string(ev.Rune())
					ui.runes = append(ui.runes, ev.Rune())
					log.Print(ui.typed)
					log.Print(ui.runes)
					ui.Draw()
					c <- &eventTyping
					continue
				} else if ev.Key() == tcell.KeyBackspace2 || ev.Key() == tcell.KeyBackspace {
					if len(ui.typed) > 0 {
						ui.typed = removeLastRune(ui.typed)
						ui.Draw()
					}
					c <- &eventTyping
					continue
				}
			}
			if ui.enableVMenu {
				switch s := ui.screen.(type) {
				case ScreenWithMenu:
					switch ev.Key() {
					case tcell.KeyDown:
						s.MenuDown()
						ui.Draw()
						continue
					case tcell.KeyUp:
						s.MenuUp()
						ui.Draw()
						continue
					case tcell.KeyEnter:
						e := s.GetMenuEvent()
						if e != nil {
							c <- e
						}
						continue
					}
				}
			}
			e := ui.findInputEvent(ev)
			if e != nil {
				log.Print("Input Event ", e.event.name)
				c <- e.event
			}
		}
	}
}

func removeLastRune(s string) string {
	_, n := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-n]
}

func (ui *UI) findInputEvent(ev *tcell.EventKey) *InputEvent {
	k, ok := keyMap[int16(ev.Key())]
	if !ok {
		return nil
	}
	m := ui.getKeysMap()
	e, ok := (*m)[k]
	if !ok {
		return nil
	}
	return &e
}

func (ui *UI) Destroy() {
	if ui != nil {
		ui.tcs.Clear()
		ui.tcs.Fini()
	}
}

func (ui *UI) DrawText(text string, style tcell.Style, cursor bool) {
	s := ui.tcs

	// w, h := s.Size()
	r := ui.emptyRow
	// c := 0

	x, y := ui.puts(style, 0, r, text)
	ui.emptyRow = y + 1
	if cursor {
		s.ShowCursor(x, y)
	}
}

func (ui *UI) DrawTextBottom(text string, style tcell.Style, cursor bool) {
	s := ui.tcs
	w, h := s.Size()

	rowsAmount := int(math.Ceil(float64(utf8.RuneCountInString(text)) / float64(w)))
	r := h - rowsAmount
	// c := 0

	x, y := ui.puts(style, 0, r, text)
	if cursor {
		s.ShowCursor(x, y)
	}
}

func (ui *UI) puts(style tcell.Style, x, y int, str string) (int, int) {
	s := ui.tcs

	w, h := s.Size()

	i := x
	var deferred []rune
	dwidth := 0
	zwj := false
	for _, r := range str {
		if y > h {
			break
		}
		if r == '\u200d' {
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
			deferred = append(deferred, r)
			zwj = true
			continue
		}
		if zwj {
			deferred = append(deferred, r)
			zwj = false
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
		if i >= w {
			y++
			i = x
		}
	}
	if len(deferred) != 0 {
		s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
	return i, y
}
