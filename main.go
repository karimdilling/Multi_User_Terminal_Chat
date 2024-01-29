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
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Could not accept connection: %v\n", err)
		}
		log.Printf("Accepted connection from %v\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	message := []byte("Hello world!\n")
	n, err := conn.Write(message)
	if err != nil {
		log.Printf("Could not write message to %v: %v\n", conn.RemoteAddr(), err)
		return
	}
	if n < len(message) {
		log.Printf("Message not complete: %d out of %d bytes written\n", n, len(message))
		return
	}
}
