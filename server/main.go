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
	defer listener.Close()
	log.Printf("Listening to TCP connections on port %v\n", PORT)
	messages := make(chan Message)
	go sendMessages(messages)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Could not accept connection: %v\n", err)
			continue
		}
		messages <- Message{conn: conn, msgType: ClientConnected}
		go receiveMessages(conn, messages)
	}
}

type MessageType int

const (
	ClientDisconnected MessageType = iota
	ClientConnected
	ClientMessage
)

type Message struct {
	conn    net.Conn
	msgType MessageType
	content string
}

func sendMessages(messages chan Message) {
	conns := []net.Conn{}

	sendMessage := func(msg Message, text string) {
		if msg.msgType == ClientMessage {
			text = msg.content
		}
		for _, conn := range conns {
			if conn == msg.conn {
				continue
			}
			_, err := conn.Write([]byte(text))
			if err != nil {
				log.Printf("Could not send message to %s: %s\n", conn.RemoteAddr(), err)
			}
		}
	}

	for {
		msg := <-messages
		switch msg.msgType {
		case ClientDisconnected:
			disconnectedAddr := msg.conn.RemoteAddr()
			for i := 0; i < len(conns); i++ {
				if (conns)[i].RemoteAddr() == disconnectedAddr {
					conns = append((conns)[:i], (conns)[i+1:]...)
					i--
				}
			}
			sendMessage(msg, "\n------- User disconnected -------\n")
			log.Printf("Client with address %s disconnected\n", msg.conn.RemoteAddr())
		case ClientConnected:
			conns = append(conns, msg.conn)
			sendMessage(msg, "\n------- New user connected -------\n")
			log.Printf("Accepted connection from %v\n", msg.conn.RemoteAddr())
		case ClientMessage:
			sendMessage(msg, "")
		}
	}
}

func receiveMessages(conn net.Conn, clientMsg chan Message) {
	buffer := make([]byte, 80)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			clientMsg <- Message{conn: conn, msgType: ClientDisconnected}
			return
		}
		content := string(buffer[0:n])
		clientMsg <- Message{conn: conn, msgType: ClientMessage, content: content}
	}
}
