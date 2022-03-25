package main

import (
	"log"
	"net"
)

type Server struct {
	address        string
	chats          []*Chat
	activeChat     int
	chatsByAddress map[string]*Chat
	conn           *net.UDPConn
}

func NewServer() *Server {
	s := Server{}
	s.chatsByAddress = make(map[string]*Chat)
	return &s
}

func (s *Server) GetActiveChat() *Chat {
	return s.chats[s.activeChat]
}

func (s *Server) GetOrCreateChat(addr string) *Chat {
	cht, ok := s.chatsByAddress[addr]
	if !ok {
		cht = NewChat(s)
		cht.SetAddress(addr)
		s.chats = append(s.chats, cht)
		s.chatsByAddress[addr] = cht
	}
	return cht
}

func (s *Server) Connect(c chan<- *Event) error {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	s.conn = conn
	defer conn.Close()
	s.address = conn.LocalAddr().String()
	c <- &eventUpdateChats

	p := make([]byte, 2048)

	for {
		n, remoteaddr, err := conn.ReadFromUDP(p)
		if err != nil {
			continue
		}
		addr := remoteaddr.String()
		msg := string(p[0:n])
		log.Print("Recieve", addr, msg)
		cht := s.GetOrCreateChat(addr)
		cht.AddMessage(msg, addr)
		c <- &eventUpdateChats
	}
}
