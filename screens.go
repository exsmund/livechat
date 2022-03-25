package main

type Screen interface {
	Draw()
}

type ScreenWithMenu interface {
	MenuUp()
	MenuDown()
	GetMenuEvent() *Event
}

type StartScreen struct {
	ui *UI
	Menu
}

func NewStartScreen(ui *UI, m *Menu) *StartScreen {
	ss := StartScreen{
		ui:   ui,
		Menu: *m,
	}
	return &ss
}

func (ss *StartScreen) Draw() {
	for i, item := range ss.Menu.items {
		style := menuItemStyle
		if i == ss.Menu.active {
			style = menuActiveItemStyle
		}
		ss.ui.DrawText(item.label, style)
	}
}

type ServerScreen struct {
	ui         *UI
	server     *Server
	menu       *Menu
	activeChat int
}

func NewServerScreen(ui *UI, s *Server, m *Menu) *ServerScreen {
	ss := ServerScreen{
		ui:     ui,
		server: s,
		menu:   m,
	}
	return &ss
}

func (ss *ServerScreen) MenuUp() {
	ss.activeChat--
	if ss.activeChat < 0 {
		ss.activeChat = len(ss.server.chats) + ss.menu.len - 1
	}
}

func (ss *ServerScreen) MenuDown() {
	ss.activeChat++
	if ss.activeChat >= len(ss.server.chats)+ss.menu.len {
		ss.activeChat = 0
	}
}

func (ss *ServerScreen) GetMenuEvent() *Event {
	if ss.activeChat < len(ss.server.chats) {
		ss.server.activeChat = ss.activeChat
		return &eventOpenChat
	}
	return ss.menu.items[ss.activeChat-len(ss.server.chats)].event
}

func (ss *ServerScreen) Draw() {
	if len(ss.server.address) > 0 {
		ss.ui.DrawText("Server: "+ss.server.address, titleStyle)
		for i, c := range ss.server.chats {
			style := menuItemStyle
			if i == ss.activeChat {
				style = menuActiveItemStyle
			}
			ss.ui.DrawText("Chat with "+c.remoteAddress, style)
		}
		chatsLen := len(ss.server.chats)
		for i, item := range ss.menu.items {
			style := menuItemStyle
			if i+chatsLen == ss.activeChat {
				style = menuActiveItemStyle
			}
			ss.ui.DrawText(item.label, style)
		}
	} else {
		ss.ui.DrawText("Creating Server...", titleStyle)
	}

}

type ConnectServerScreen struct {
	ui *UI
}

func NewConnectServerScreen(ui *UI) *ConnectServerScreen {
	css := ConnectServerScreen{
		ui: ui,
	}
	return &css
}

func (ss *ConnectServerScreen) Draw() {
	ss.ui.DrawText("Input server address", titleStyle)
	ss.ui.DrawText(ss.ui.typed, inputStyle)
}

type ChatScreen struct {
	ui   *UI
	chat *Chat
}

func NewChatScreen(ui *UI, c *Chat) *ChatScreen {
	cs := ChatScreen{
		ui:   ui,
		chat: c,
	}
	return &cs
}

func (cs *ChatScreen) Draw() {
	cs.ui.DrawText("Chat with "+cs.chat.remoteAddress, titleStyle)

	for _, m := range cs.chat.messages {
		style := receivedMsgStyle
		if m.sender == cs.chat.server.address {
			style = myMsgStyle
		}
		cs.ui.DrawText(m.text, style)
	}

	cs.ui.DrawTextBottom(cs.ui.typed, inputStyle)
}
