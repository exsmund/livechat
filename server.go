package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/ccding/go-stun/stun"
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
	// err := s.connectWithStun()
	conn, err := s.connectLocal()
	if err != nil {
		return err
	}
	defer conn.Close()
	s.conn = conn
	c <- &eventUpdateChats

	p := make([]byte, 2048)

	for {
		n, remoteaddr, err := s.conn.ReadFromUDP(p)
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

func (s *Server) connectWithStun() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := stun.NewClientWithConnection(conn)
	client.SetServerAddr("stun.sonetel.com:3478")
	client.Keepalive()
	nat, host, err := client.Discover()
	log.Println("NAT Type:", nat)
	if err != nil {
		return nil, err
	}
	if host != nil {
		log.Println("External IP Family:", host.Family())
		log.Println("External IP:", host.IP())
		log.Println("External Port:", host.Port())
		s.address = host.IP() + ":" + fmt.Sprint(host.Port())
	} else {
		return nil, errors.New("some error")
	}
	return conn, nil
}

func (s *Server) connectLocal() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	s.conn = conn
	s.address = conn.LocalAddr().String()
	return conn, nil
}
