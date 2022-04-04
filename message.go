package main

import (
	"time"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

type Message struct {
	text     string
	order    uint
	ts       time.Time
	sender   string
	finished bool
	own      bool
	runes    [][]rune
}

func (m *Message) SetText(t string) {
	m.text = t
	m.updateRunes()
}

func (m *Message) updateRunes() {
	runes := make([][]rune, utf8.RuneCountInString(m.text))
	i := 0
	var deferred []rune
	dwidth := 0
	zwj := false
	addToLine := func(r []rune, runeWidth int) {
		if len(r) != 0 {
			runes[i] = r
			i += runeWidth
		}
	}
	for _, r := range m.text {
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
			addToLine(deferred, dwidth)
			deferred = nil
			dwidth = 1
		case 2:
			addToLine(deferred, dwidth)
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	addToLine(deferred, dwidth)
}

func NewMessage(
	text string,
	order uint,
	ts time.Time,
	sender string,
	finished bool,
	own bool,
) *Message {
	m := Message{
		text:     text,
		order:    order,
		ts:       ts,
		sender:   sender,
		finished: finished,
		own:      own,
	}
	m.updateRunes()
	return &m
}
