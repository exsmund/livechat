package main

import "log"

type App struct {
	state       AppState
	inputEvents chan *Event
	ui          *UI
	server      *Server
	activeChat  *Chat
}

func NewApp() (*App, error) {
	a := App{}
	err := a.init()
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (a *App) init() error {
	var err error

	defer func() {
		if err != nil {
			a.Destroy()
		}
	}()

	a.inputEvents = make(chan *Event)
	a.setState(appStateStarting)

	a.ui, err = NewUI(a.getKeys, a.inputEvents)
	if err != nil {
		return err
	}
	a.ui.SetScreen(NewStartScreen(a.ui, &startMenu), false, true)

	go func() {
		if e := a.ui.Run(); e != nil {
			log.Fatal(e.Error())
		}
	}()

	return nil
}

func (a *App) Loop() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Recovered in f", r)
		}
		a.Destroy()
	}()
	log.Print("Loop")

	// go a.ui.Listen(a.inputEvents)
	a.drawUI()
	for a.state != appStateEnded {
		select {
		case e := <-a.inputEvents:
			log.Print("Dispatch Input Event ", e)
			a.dispatchEvent(e)
		}
	}
	// a.ui.Refresh()
	// a.ui.Quit()
	log.Print("End loop")
}

func (a *App) Destroy() {
	if a.ui != nil {
		a.ui.Destroy()
	}
}

func (a *App) getKeys() *KeyEventMap {
	m := stateEventMap[a.state]
	return &m
}

func (a *App) drawUI() {
	a.ui.Update2()
}

func (a *App) setState(s AppState) {
	a.state = s
}

func (a *App) dispatchEvent(e *Event) {
	switch e {
	case &eventDestroy:
		a.setState(appStateEnded)
	case &eventCreateServer:
		a.createServer()
	case &eventCreateChat:
		a.createChat()
	case &eventConnectServer:
		a.connectServer()
	case &eventOpenChat:
		a.openChat()
	case &eventSendMessage:
		a.sendMessage()
	case &eventTyping:
		a.typing()
	case &eventBack:
		a.eventBack()
	default:
		a.drawUI()
	}
}

func (a *App) createServer() {
	a.server = NewServer()
	a.ui.SetScreen(NewServerScreen(a.ui, a.server, &serverMenu), false, true)
	a.setState(appStateServer)
	a.drawUI()
	go a.server.Connect(a.inputEvents)
}

func (a *App) createChat() {
	a.ui.SetScreen(NewConnectServerScreen(a.ui), true, false)
	a.setState(appStateNewChat)
	a.drawUI()
}

func (a *App) connectServer() {
	c := a.server.GetOrCreateChat(a.ui.typed)
	a.activeChat = c
	a.ui.SetScreen(NewChatScreen(a.ui, c), true, false)
	a.setState(appStateChat)
	a.drawUI()
}

func (a *App) openChat() {
	c := a.server.GetActiveChat()
	a.activeChat = c
	a.ui.SetScreen(NewChatScreen(a.ui, c), true, false)
	a.setState(appStateChat)
	a.drawUI()
}

func (a *App) typing() {
	switch a.state {
	case appStateChat:
		a.activeChat.Typing(a.ui.typed)
	}
}

func (a *App) sendMessage() {
	if a.activeChat != nil {
		a.activeChat.Send(a.ui.typed)
		a.ui.ClearTyped()
		a.drawUI()
	}
}

func (a *App) eventBack() {
	switch a.state {
	case appStateServer:
		a.ui.SetScreen(NewStartScreen(a.ui, &startMenu), false, true)
		a.setState(appStateStarting)
		a.drawUI()
	case appStateNewChat, appStateChat:
		a.ui.SetScreen(NewServerScreen(a.ui, a.server, &serverMenu), false, true)
		a.setState(appStateServer)
		a.drawUI()
	}
}
