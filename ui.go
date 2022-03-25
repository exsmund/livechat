package main

import (
	"log"
	"math"

	"github.com/gdamore/tcell/v2"
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
		switch ev := ev.(type) {
		case *tcell.EventResize:
			ui.tcs.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				panic("err")
			}
			log.Print("Input |", ev.Name(), "|", ev.Key())
			if ui.enableTyping {
				if ev.Key() == tcell.KeyRune {
					ui.typed += string(ev.Rune())
					ui.Draw()
					continue
				} else if ev.Key() == tcell.KeyBackspace2 || ev.Key() == tcell.KeyBackspace {
					if len(ui.typed) > 0 {
						ui.typed = ui.typed[:len(ui.typed)-1]
						ui.Draw()
					}
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

func (ui *UI) DrawText(text string, style tcell.Style) {
	s := ui.tcs
	w, h := s.Size()
	r := ui.emptyRow
	c := 0

	for _, ru := range []rune(text) {
		if r > h {
			break
		}
		s.SetContent(c, r, ru, nil, style)
		c++
		if c >= w {
			r++
			c = 0
		}
	}
	ui.emptyRow = r + 1
}

func (ui *UI) DrawTextBottom(text string, style tcell.Style) {
	s := ui.tcs
	w, h := s.Size()

	rowsAmount := int(math.Ceil(float64(len(text)) / float64(w)))
	r := h - rowsAmount
	c := 0

	for _, ru := range []rune(text) {
		if r > h {
			break
		}
		s.SetContent(c, r, ru, nil, style)
		c++
		if c >= w {
			r++
			c = 0
		}
	}
}
