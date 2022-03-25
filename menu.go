package main

type MenuItem struct {
	label string
	event *Event
}

type Menu struct {
	items  []MenuItem
	active int
	len    int
}

func (m *Menu) MenuUp() {
	m.active--
	if m.active < 0 {
		m.active = m.len - 1
	}
}

func (m *Menu) MenuDown() {
	m.active++
	if m.active >= m.len {
		m.active = 0
	}
}

func (m *Menu) GetMenuEvent() *Event {
	return m.items[m.active].event
}
