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
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go readMessages(conn, &wg)
	go sendMessage(conn)
	wg.Wait()
}

func readMessages(conn net.Conn, wg *sync.WaitGroup) {
	rdbuff := make([]byte, 80)
	for {
		n, err := conn.Read(rdbuff)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Connection closed by foreign host.")
			conn.Close()
			wg.Done()
			return
		}
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
