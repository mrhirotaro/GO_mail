package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Session構造体を定義して、クライアントごとの状態を管理
type Session struct {
	Domain        string
	SenderMail    string
	RecipientMail []string // 複数受信者を保持するためにスライスに変更
}

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
	session := &Session{}

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
			handleHelo(conn, parts, session)
			fmt.Println("Session after HELO/EHLO:", session)
		} else if len(parts) >= 1 && parts[0] == "MAIL" && strings.HasPrefix(parts[1], "FROM:") {
			handleMailFrom(conn, parts, session)
			fmt.Println("Session after MAIL FROM:", session)

		} else if len(parts) >= 1 && parts[0] == "RCPT" && strings.HasPrefix(parts[1], "TO:") {
			handleMailRcpt(conn, parts, session)
			fmt.Println("Session after RCPT TO:", session)

		} else if len(parts) >= 1 && parts[0] == "RSET" {
			session = &Session{}
			handleRset(conn)
			fmt.Println("Session after RSET:", session)

		} else if len(parts) >= 1 && parts[0] == "QUIT" {
			sendResponse(conn, "221 Bye\r\n")
			return
		} else {
			sendResponse(conn, "502 Unrecognized command\r\n")
		}
	}
}

func handleHelo(conn net.Conn, parts []string, session *Session) {
	if len(parts) >= 2 {
		domain := strings.ToLower(parts[1])
		session.Domain = domain
		sendResponse(conn, "250 Hello "+domain+"\r\n")
	} else {
		sendResponse(conn, "501 Syntax error in parameters or arguments\r\n")
	}
}

func handleMailFrom(conn net.Conn, parts []string, session *Session) {
	senderMail := strings.TrimPrefix(parts[1], "FROM:")
	if senderMail == "" && len(parts) >= 2 {
		senderMail = parts[2]
	}

	senderMail = strings.Trim(senderMail, "<>")

	if senderMail != "" {
		session.SenderMail = senderMail
		sendResponse(conn, "250 OK\r\n")
	} else {
		sendResponse(conn, "501 Syntax error in parameters or arguments\r\n")
	}
}
func handleMailRcpt(conn net.Conn, parts []string, session *Session) {
	recipientMail := strings.TrimPrefix(parts[1], "TO:")
	if recipientMail == "" && len(parts) >= 2 {
		recipientMail = parts[2]
	}

	recipientMail = strings.Trim(recipientMail, "<>")

	if recipientMail != "" {
		session.RecipientMail = append(session.RecipientMail, recipientMail)
		sendResponse(conn, "250 OK\r\n")
	} else {
		sendResponse(conn, "501 Syntax error in parameters or arguments\r\n")
	}
}

func handleRset(conn net.Conn) {
	sendResponse(conn, "250 OK\r\n")
}

func sendResponse(conn net.Conn, message string) {
	fmt.Println("send", message)
	conn.Write([]byte(message))
}
