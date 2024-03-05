package main

import (
	"encoding/json"
	"fmt"
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
	clients := map[string]string{} // maps IP-address:port to username
	go sendMessages(messages, clients)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Could not accept connection: %v\n", err)
			continue
		}
		go receiveMessages(conn, messages, clients)
	}
}

type MessageType int

const (
	ClientDisconnected MessageType = iota
	ClientConnected
	ClientMessage
)

type Message struct {
	conn       net.Conn
	Username   string      `json:"username"`
	MsgType    MessageType `json:"msg_type"`
	Content    string      `json:"content"`
	ClientList []string    `json:"client_list"`
}

func sendMessages(messages chan Message, clients map[string]string) {
	conns := []net.Conn{}
	usernames := []string{}

	sendMessage := func(msg *Message, text string) {
		if msg.MsgType == ClientMessage {
			text = msg.Content
		}
		for _, conn := range conns {
			if conn == msg.conn && msg.MsgType != ClientConnected {
				continue
			}
			msg.Content = text
			msgJSON, err := json.Marshal(&msg)
			if err != nil {
				log.Printf("Could not parse JSON message: %v\n", err)
			}
			_, err = conn.Write([]byte(msgJSON))
			if err != nil {
				log.Printf("Could not send message to %s: %s\n", conn.RemoteAddr(), err)
			}
		}
	}

	for {
		msg := <-messages
		switch msg.MsgType {
		case ClientDisconnected:
			disconnectedAddr := msg.conn.RemoteAddr()
			for i := 0; i < len(conns); i++ {
				if (conns)[i].RemoteAddr() == disconnectedAddr {
					conns = append((conns)[:i], (conns)[i+1:]...)
					usernames = append(usernames[:i], usernames[i+1:]...)
					i--
				}
			}
			delete(clients, disconnectedAddr.String())
			msg.ClientList = usernames
			sendMessage(&msg, fmt.Sprintf("\n------- %s disconnected -------\n", msg.Username))
			log.Printf("Client with address %s disconnected\n", msg.conn.RemoteAddr())
		case ClientConnected:
			conns = append(conns, msg.conn)
			clients[msg.conn.RemoteAddr().String()] = msg.Username
			usernames = append(usernames, msg.Username)
			msg.ClientList = usernames
			sendMessage(&msg, fmt.Sprintf("\n------- %s just connected -------\n", msg.Username))
			log.Printf("Accepted connection from %v\n", msg.conn.RemoteAddr())
		case ClientMessage:
			sendMessage(&msg, "")
		}
	}
}

func receiveMessages(conn net.Conn, messages chan Message, clients map[string]string) {
	buffer := make([]byte, 250)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			messages <- Message{conn: conn, Username: clients[conn.RemoteAddr().String()], MsgType: ClientDisconnected}
			return
		}
		var msg Message
		json.Unmarshal(buffer[0:n], &msg)
		messages <- Message{
			conn:     conn,
			Username: msg.Username,
			MsgType:  msg.MsgType,
			Content:  msg.Content,
		}
	}
}
