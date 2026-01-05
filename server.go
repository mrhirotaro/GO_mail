package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:1025")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on localhost:1025")

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Client connected:", conn.RemoteAddr())

	// Handle client connection here
	handleClient(conn)

}

func handleClient(conn net.Conn) {
	redeader := bufio.NewReader(conn)

	sendResponse(conn, "220 Welcome to the SMTP server\r\n")

	for {
		line, err := redeader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from client:", err)
			return
		}

		command := strings.TrimSpace(line)
		fmt.Println("recv", command)
		commandUpper := strings.ToUpper(command)

		parts := strings.Fields(commandUpper)
		if len(parts) >= 1 && (parts[0] == "HELO" || parts[0] == "EHLO") {
			if len(parts) >= 2 {
				domain := strings.ToLower(parts[1])
				sendResponse(conn, "250 Hello "+domain+"\r\n")
			} else {
				sendResponse(conn, "501 Syntax error in parameters or arguments\r\n")
			}
		} else if len(parts) >= 1 && parts[0] == "QUIT" {
			sendResponse(conn, "221 Bye\r\n")
			return
		} else {
			sendResponse(conn, "502 Unrecognized command\r\n")
		}
	}
}

func sendResponse(conn net.Conn, message string) {
	fmt.Println("send", message)
	conn.Write([]byte(message))
}
