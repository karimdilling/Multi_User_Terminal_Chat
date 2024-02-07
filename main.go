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
		messages <- Message{conn, ""}
		go receiveMessages(conn, messages)
	}
}

type Message struct {
	conn    net.Conn
	content string
}

func sendMessages(messages chan Message, conns *[]net.Conn) {
	for {
		msg := <-messages
		for _, conn := range *conns {
			if conn == msg.conn {
				continue
			}
			_, err := conn.Write([]byte(msg.content))
			if err != nil {
				log.Printf("Could not send message to %s: %s", conn.RemoteAddr(), err)
			}
		}
	}
}

func receiveMessages(conn net.Conn, clientMsg chan Message) {
	defer conn.Close()

	for {
		buffer := make([]byte, 200)
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			return
		}
		content := string(buffer[0:n])
		clientMsg <- Message{conn, content}
	}
}
