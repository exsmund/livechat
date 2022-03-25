package main

import (
	"log"
	"net"
	"time"
)

type Message struct {
	text   string
	ts     time.Time
	sender string
}

type Chat struct {
	remoteAddress string
	updAddr       *net.UDPAddr
	messages      []Message
	server        *Server
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

func (c *Chat) AddMessage(msg string, sender string) {
	c.messages = append(c.messages, Message{msg, time.Now(), sender})
}

// func (c *Chat) Connect() {
// 	conn, err := net.Dial("udp", c.remoteAddress)
// 	if err != nil {
// 		log.Printf("Some error %v", err)
// 		return
// 	}
// 	c.conn = &conn
// }

// func (c *Chat) Disconnect() {
// 	if c.conn != nil {
// 		(*c.conn).Close()
// 		c.conn = nil
// 	}
// }

func (c *Chat) Send(msg string) {
	log.Print("Send msg ", msg, " to ", c.updAddr)
	c.server.conn.WriteToUDP([]byte(msg), c.updAddr)
	// fmt.Fprintf(*c.server.conn, msg)
	c.AddMessage(msg, c.server.address)
}
