package main

type Event struct {
	name string
}

type EventMessage struct {
	name string
}

type InputEvent struct {
	event *Event
}

type KeyEventMap map[string]InputEvent
