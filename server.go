package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/ccding/go-stun/stun"
	"github.com/pion/turn/v2"
)

type PackedMsg struct {
	Msg      string
	Order    uint
	Finished bool
	Addr     string
}

type Server struct {
	address        string
	chats          []*Chat
	activeChat     int
	chatsByAddress map[string]*Chat
	conn           net.PacketConn
	relayConn      net.PacketConn
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
	// err := s.connectLocal(c)
	err := s.connectWithTurn(c)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (s *Server) Send(msg string, o uint, f bool, addr *net.UDPAddr) error {
	p := &PackedMsg{msg, o, f, s.address}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Print("send to ", addr, " from ", s.conn.LocalAddr().String())
	s.conn.WriteTo(buf.Bytes(), addr)
	return nil
}

func (s *Server) connectWithStun(c chan<- *Event) error {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := stun.NewClientWithConnection(conn)
	client.SetServerAddr("157.230.107.39:3478")
	client.Keepalive()
	nat, host, err := client.Discover()
	log.Println("NAT Type:", nat)
	if err != nil {
		return err
	}
	if host != nil {
		log.Println("External IP Family:", host.Family())
		log.Println("External IP:", host.IP())
		log.Println("External Port:", host.Port())
		s.address = host.IP() + ":" + fmt.Sprint(host.Port())
	} else {
		return errors.New("some error")
	}

	s.conn = conn
	c <- &eventUpdateChats

	s.listen(conn, c)
	return nil
}

func (s *Server) connectLocal(c chan<- *Event) error {
	addr, err := net.ResolveUDPAddr("udp", "localhost:0")
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	s.conn = conn
	s.address = conn.LocalAddr().String()
	c <- &eventUpdateChats

	s.listen(conn, c)
	return nil
}

func (s *Server) connectWithTurn(c chan<- *Event) error {
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()

	cfg := &turn.ClientConfig{
		Conn:           conn,
		STUNServerAddr: "openrelay.metered.ca:80",
		TURNServerAddr: "openrelay.metered.ca:80",
		Username:       "openrelayproject",
		Password:       "openrelayproject",
	}

	client, err := turn.NewClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.Listen()
	if err != nil {
		return err
	}

	relayConn, err := client.Allocate()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := relayConn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()

	log.Printf("relayed-address=%s", relayConn.LocalAddr().String())

	mappedAddr, err := client.SendBindingRequest()
	if err != nil {
		return err
	}
	log.Printf("mapped-address=%s", mappedAddr.String())

	_, err = relayConn.WriteTo([]byte("Hello"), mappedAddr)
	if err != nil {
		return err
	}

	s.address = relayConn.LocalAddr().String()
	s.conn = conn
	s.relayConn = relayConn
	c <- &eventUpdateChats
	s.listen(relayConn, c)
	return nil
}

func (s *Server) listen(conn net.PacketConn, c chan<- *Event) error {
	p := make([]byte, 2048)

	for {
		n, remoteaddr, err := conn.ReadFrom(p)
		log.Print("Recieve", remoteaddr, p[:n])
		if err != nil {
			continue
		}
		buf := bytes.NewBuffer(p[0:n])
		dec := gob.NewDecoder(buf)
		var p PackedMsg
		err = dec.Decode(&p)
		if err != nil {
			continue
		}

		addr := remoteaddr.String()
		log.Print("Recieve", addr, p)
		cht := s.GetOrCreateChat(p.Addr)
		cht.AddReceivedMessage(p, addr)
		c <- &eventUpdateChats
	}
}
