package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Screen interface {
	Update()
	views.Widget
}

type ScreenWithMenu interface {
	MenuUp()
	MenuDown()
	GetMenuEvent() *Event
}

type StartScreen struct {
	ui      *UI
	widgets []*views.TextBar
	Menu
	views.BoxLayout
}

func NewStartScreen(ui *UI, m *Menu) *StartScreen {
	s := StartScreen{
		ui:   ui,
		Menu: *m,
	}
	s.SetOrientation(views.Vertical)
	for i, item := range m.items {
		t := views.NewTextBar()
		style := menuItemStyle
		if i == s.Menu.active {
			style = menuActiveItemStyle
		}
		t.SetCenter(item.label, style)
		s.AddWidget(t, 0)
		s.widgets = append(s.widgets, t)
	}

	return &s
}

func (ss *StartScreen) HandleEvent(ev tcell.Event) bool {
	return ss.ui.HandleEvent(ev)
}

func (ss *StartScreen) Update() {
	for i, item := range ss.Menu.items {
		style := menuItemStyle
		if i == ss.Menu.active {
			style = menuActiveItemStyle
		}
		ss.widgets[i].SetCenter(item.label, style)
	}
}

type ServerScreen struct {
	ui     *UI
	server *Server
	Menu
	activeChat int
	top        *views.TextBar
	chatWidget *views.BoxLayout
	widgets    []*views.TextBar
	views.BoxLayout
}

func NewServerScreen(ui *UI, s *Server, m *Menu) *ServerScreen {
	ss := ServerScreen{
		ui:         ui,
		server:     s,
		Menu:       *m,
		top:        views.NewTextBar(),
		chatWidget: views.NewBoxLayout(views.Vertical),
	}
	ss.top.SetCenter("Creating Server...", titleStyle)
	ss.AddWidget(ss.top, 0)
	ss.AddWidget(ss.chatWidget, 0)
	ss.SetOrientation(views.Vertical)

	for i, item := range m.items {
		t := views.NewTextBar()
		style := menuItemStyle
		if i == ss.Menu.active {
			style = menuActiveItemStyle
		}
		t.SetCenter(item.label, style)
		ss.AddWidget(t, 0)
		ss.widgets = append(ss.widgets, t)
	}
	return &ss
}

func (s *ServerScreen) HandleEvent(ev tcell.Event) bool {
	return s.ui.HandleEvent(ev)
}

func (ss *ServerScreen) MenuUp() {
	ss.activeChat--
	if ss.activeChat < 0 {
		ss.activeChat = len(ss.server.chats) + ss.Menu.len - 1
	}
}

func (ss *ServerScreen) MenuDown() {
	ss.activeChat++
	if ss.activeChat >= len(ss.server.chats)+ss.Menu.len {
		ss.activeChat = 0
	}
}

func (ss *ServerScreen) GetMenuEvent() *Event {
	if ss.activeChat < len(ss.server.chats) {
		ss.server.activeChat = ss.activeChat
		return &eventOpenChat
	}
	return ss.Menu.items[ss.activeChat-len(ss.server.chats)].event
}

func (ss *ServerScreen) Update() {
	log.Print("ServerScreen update")
	if len(ss.server.address) > 0 {
		ss.top.SetCenter("Server: "+ss.server.address, titleStyle)
		for i, c := range ss.server.chats {
			style := menuItemStyle
			if i == ss.activeChat {
				style = menuActiveItemStyle
			}
			ss.ui.DrawText("Chat with "+c.remoteAddress, style, false)
		}
		chatsLen := len(ss.server.chats)
		for i, item := range ss.Menu.items {
			style := menuItemStyle
			if i+chatsLen == ss.activeChat {
				style = menuActiveItemStyle
			}
			ss.widgets[i].SetCenter(item.label, style)
		}
	} else {
		ss.top.SetCenter("Creating Server...", titleStyle)
	}

}

type ConnectServerScreen struct {
	ui    *UI
	top   *views.TextBar
	input *views.TextBar

	views.BoxLayout
}

func NewConnectServerScreen(ui *UI) *ConnectServerScreen {
	s := ConnectServerScreen{
		ui:    ui,
		top:   views.NewTextBar(),
		input: views.NewTextBar(),
	}
	s.SetOrientation(views.Vertical)
	s.top.SetCenter("Input server address", titleStyle)
	s.AddWidget(s.top, 0)
	// s.input.SetStyle(inputStyle)
	s.AddWidget(s.input, 0)
	return &s
}

func (ss *ConnectServerScreen) Update() {
	// ss.ui.DrawText("Input server address", titleStyle, false)
	// ss.ui.DrawText(ss.ui.typed, inputStyle, true)
	ss.input.SetCenter(ss.ui.typed, inputStyle)
}

type ChatScreen struct {
	ui   *UI
	chat *Chat
	views.BoxLayout
}

func NewChatScreen(ui *UI, c *Chat) *ChatScreen {
	cs := ChatScreen{
		ui:   ui,
		chat: c,
	}
	return &cs
}

func (cs *ChatScreen) Update() {
	cs.ui.DrawText("Chat with "+cs.chat.remoteAddress, titleStyle, false)
	oi := uint(0)
	ri := uint(0)
	for {
		if oi+1 > cs.chat.amountOwnMsgs && ri+1 > cs.chat.amountReceivedMsgs {
			break
		} else {
			if oi+1 > cs.chat.amountOwnMsgs {
				rm := &cs.chat.receivedMessages[ri]
				cs.drawMsg(rm, receivedMsgStyle)
				ri++
			} else if ri+1 > cs.chat.amountReceivedMsgs {
				om := &cs.chat.ownMessages[oi]
				cs.drawMsg(om, myMsgStyle)
				oi++
			} else {
				om := &cs.chat.ownMessages[oi]
				rm := &cs.chat.receivedMessages[ri]
				if rm.ts.Before(om.ts) {
					cs.drawMsg(rm, receivedMsgStyle)
					ri++
				} else {
					cs.drawMsg(om, myMsgStyle)
					oi++
				}
			}
		}
	}

	cs.ui.DrawTextBottom(cs.ui.typed, inputStyle, true)
}

func (cs *ChatScreen) drawMsg(m *Message, style tcell.Style) {
	msg := m.text
	if !m.finished {
		msg = msg + "..."
	}
	cs.ui.DrawText(msg, style, false)
}
