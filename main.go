package main

import (
	"log"
	"net"
)

func main() {
	const PORT = "8080"
	listener, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatalf("Could not listen on port %v: %v\n", PORT, err)
	}
	log.Printf("Listening to TCP connections on port %v\n", PORT)
	messages := make(chan Message)
	conns := []net.Conn{}
	go sendMessages(messages, &conns)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Could not accept connection: %v\n", err)
			continue
		}
		conns = append(conns, conn)
		log.Printf("Accepted connection from %v\n", conn.RemoteAddr())
		go receiveMessages(conn, messages)
	}
}

type MessageType int

const (
	ClientDisconnected MessageType = iota
	ClientMessage
)

type Message struct {
	conn    net.Conn
	msgType MessageType
	content string
}

func sendMessages(messages chan Message, conns *[]net.Conn) {
	for {
		msg := <-messages
		if msg.msgType == ClientDisconnected {
			disconnectedAddr := msg.conn.RemoteAddr()
			for i := 0; i < len(*conns); i++ {
				if (*conns)[i].RemoteAddr() == disconnectedAddr {
					*conns = append((*conns)[:i], (*conns)[i+1:]...)
					i--
				}
			}
			log.Printf("Client with address %s disconnected\n", msg.conn.RemoteAddr())
		}
		for _, conn := range *conns {
			if conn == msg.conn {
				continue
			}
			_, err := conn.Write([]byte(msg.content))
			if err != nil {
				log.Printf("Could not send message to %s: %s\n", conn.RemoteAddr(), err)
			}
		}
	}
}

func receiveMessages(conn net.Conn, clientMsg chan Message) {
	defer conn.Close()

	buffer := make([]byte, 80)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			clientMsg <- Message{conn, ClientDisconnected, ""}
			return
		}
		content := string(buffer[0:n])
		clientMsg <- Message{conn, ClientMessage, content}
	}
}
