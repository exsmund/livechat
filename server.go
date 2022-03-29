package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/ccding/go-stun/stun"
	"github.com/pion/logging"
	"github.com/pion/turn/v2"
)

type PackedMsg struct {
	Msg      string
	Order    uint
	Finished bool
}

type Server struct {
	address        string
	chats          []*Chat
	activeChat     int
	chatsByAddress map[string]*Chat
	conn           net.PacketConn
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
	// conn, err := s.connectWithStun()
	conn, err := s.connectWithTurn(c)
	if err != nil {
		log.Fatal(err)
		return err
	}
	s.conn = conn
	defer s.conn.Close()
	c <- &eventUpdateChats

	p := make([]byte, 2048)

	for {
		n, remoteaddr, err := conn.ReadFrom(p)
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
		cht := s.GetOrCreateChat(addr)
		cht.AddReceivedMessage(p, addr)
		c <- &eventUpdateChats
	}
}

func (s *Server) Send(p *PackedMsg, addr *net.UDPAddr) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
		return err
	}

	s.conn.WriteTo(buf.Bytes(), addr)
	return nil
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
	client.SetServerAddr("157.230.107.39:3478")
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

func (s *Server) connectWithTurn(c chan<- *Event) (net.PacketConn, error) {
	// addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	// if err != nil {
	// 	return nil, err
	// }
	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	realm := "exs"
	cfg := &turn.ClientConfig{
		// STUNServerAddr: "157.230.107.39:6473",
		// TURNServerAddr: "157.230.107.39:6473",
		STUNServerAddr: "numb.viagenie.ca:3478",
		TURNServerAddr: "numb.viagenie.ca:3478",
		Conn:           conn,
		Username:       "exsmund@gmail.com",
		Password:       "8sCH9NnBtL9Bi8dMWhpe",
		Realm:          realm,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	err = client.Listen()
	if err != nil {
		return nil, err
	}

	relayConn, err := client.Allocate()
	if err != nil {
		return nil, err
	}

	log.Printf("relayed-address=%s", relayConn.LocalAddr().String())

	mappedAddr, err := client.SendBindingRequest()
	if err != nil {
		return nil, err
	}
	log.Printf("mapped-address=%s", mappedAddr.String())

	_, err = relayConn.WriteTo([]byte("Hello"), mappedAddr)
	if err != nil {
		return nil, err
	}

	s.address = relayConn.LocalAddr().String()
	// c <- &eventUpdateChats

	// buf := make([]byte, 1600)
	// for {
	// 	n, from, readerErr := relayConn.ReadFrom(buf)
	// 	if readerErr != nil {
	// 		break
	// 	}
	// 	log.Print(buf[:n], from)

	// 	// Echo back
	// 	if _, readerErr = relayConn.WriteTo(buf[:n], from); readerErr != nil {
	// 		break
	// 	}
	// }

	return relayConn, nil
}
