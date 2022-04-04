package main

import (
	"log"
	"net"
	"time"
)

type Chat struct {
	remoteAddress      string
	updAddr            *net.UDPAddr
	allMessages        []*Message
	ownMessages        []*Message
	amountOwnMsgs      uint
	receivedMessages   []*Message
	amountReceivedMsgs uint
	server             *Server
	// conn          *net.Conn
}

func NewChat(s *Server) *Chat {
	c := Chat{server: s}
	return &c
}

func (c *Chat) SetAddress(address string) {
	c.remoteAddress = address
	addr, err := net.ResolveUDPAddr("udp", address)
	if err == nil {
		c.updAddr = addr
	} else {
		c.updAddr = nil
	}
}

func (c *Chat) AddOwnMessage(msg string, sender string) {
	m := NewMessage(
		msg,
		c.amountOwnMsgs,
		time.Now(),
		sender,
		true,
		true,
	)
	c.ownMessages = append(c.ownMessages, m)
	c.allMessages = append(c.allMessages, m)
	c.amountOwnMsgs++
}

func (c *Chat) AddReceivedMessage(p PackedMsg, sender string) {
	log.Print("Add msg", p.Msg, p.Order, c.amountReceivedMsgs)
	if c.amountReceivedMsgs < p.Order+1 {
		for i := c.amountReceivedMsgs; i <= p.Order; i++ {
			var m *Message
			if i == p.Order {
				m = NewMessage(
					p.Msg,
					i,
					time.Now(),
					sender,
					p.Finished,
					false,
				)
			} else {
				m = NewMessage(
					"",
					i,
					time.Now(),
					sender,
					p.Finished,
					false,
				)
			}
			c.receivedMessages = append(c.receivedMessages, m)
			c.allMessages = append(c.allMessages, m)
		}
		c.amountReceivedMsgs = p.Order + 1
	} else {
		c.receivedMessages[p.Order].SetText(p.Msg)
		c.receivedMessages[p.Order].ts = time.Now()
		c.receivedMessages[p.Order].finished = p.Finished
	}
}

func (c *Chat) Typing(msg string) {
	c.server.Send(msg, c.amountOwnMsgs, false, c.updAddr)
}

func (c *Chat) Send(msg string) {
	c.AddOwnMessage(msg, c.server.address)
	c.server.Send(msg, c.amountOwnMsgs-1, true, c.updAddr)
}
