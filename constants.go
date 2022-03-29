package main

type AppState int

const (
	appStateStarting AppState = 0
	appStateEnded    AppState = 1
	appStateServer   AppState = 2
	appStateNewChat  AppState = 3
	appStateChat     AppState = 4
)

// App Events
var (
	eventDestroy       = Event{"eventDestroy"}
	eventMenuUp        = Event{"menuUp"}
	eventMenuDown      = Event{"menuDown"}
	eventMenuSelect    = Event{"menuSelect"}
	eventCreateServer  = Event{"createServer"}
	eventCreateChat    = Event{"createChat"}
	eventConnectServer = Event{"connectServer"}
	eventUpdateChats   = Event{"updateChats"}
	eventBack          = Event{"back"}
	eventTyping        = Event{"typing"}
	eventSendMessage   = Event{"sendMessage"}
	eventOpenChat      = Event{"openChat"}
)

var stateEventMap = map[AppState]KeyEventMap{
	appStateStarting: {
		"Enter": {
			event: &eventMenuSelect,
		},
		"Up": {
			event: &eventMenuUp,
		},
		"Down": {
			event: &eventMenuDown,
		},
		"Esc": {
			event: &eventDestroy,
		},
	},
	appStateNewChat: {
		"Enter": {
			event: &eventConnectServer,
		},
		"Esc": {
			event: &eventBack,
		},
	},
	appStateChat: {
		"Enter": {
			event: &eventSendMessage,
		},
		"Esc": {
			event: &eventBack,
		},
	},
	appStateServer: {
		"Esc": {
			event: &eventBack,
		},
	},
}

var startMenu = Menu{
	items: []MenuItem{
		{
			"Start chatting",
			&eventCreateServer,
		},
		{
			"Exit",
			&eventDestroy,
		},
	},
	len: 2,
}

var serverMenu = Menu{
	items: []MenuItem{
		{
			"New chat",
			&eventCreateChat,
		},
		{
			"Stop chating",
			&eventBack,
		},
	},
	len: 2,
}
