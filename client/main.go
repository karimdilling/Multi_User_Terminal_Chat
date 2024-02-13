package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to server.")
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go readMessages(conn)
	go sendMessage(conn)
	wg.Wait()
}

func readMessages(conn net.Conn) {
	rdbuff := make([]byte, 80)
	for {
		n, _ := conn.Read(rdbuff)
		fmt.Println(string(rdbuff[0:n]))
	}
}

func sendMessage(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		conn.Write([]byte(input))
	}
}
