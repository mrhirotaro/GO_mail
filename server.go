package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SMTPServer struct {
	Port       int
	listener   net.Listener
	DomainRuts map[string]string
}

// Session構造体を定義して、クライアントごとの状態を管理
type Session struct {
	Domain        string
	SenderMail    string
	RecipientMail []string // 複数受信者を保持するためにスライス
	Body          string   // メール本文を保持
}

func NewSMTPServer(port int, domainRoutes map[string]string) *SMTPServer {
	return &SMTPServer{
		Port:       port,
		DomainRuts: domainRoutes,
	}
}

func (s *SMTPServer) Start(ctx context.Context) error { // 他所からアクセスできるようにするために大文字で始める
	// addressの文字列を作成
	address := fmt.Sprintf("localhost:%d", s.Port)
	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return err
	}
	defer s.listener.Close()
	//fmt.Println("Server listening on localhost:", s.Port)
	fmt.Println("Server listening on", address)
	//fmt.Println("Domain Routes:", s.DomainRuts)

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down server...")
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				fmt.Println("Server has been shut down.")
				return nil
			default:
				fmt.Println("Error accepting connection:", err)
				return err
			}
		}
		fmt.Println("Client connected:", conn.RemoteAddr())

		// Handle client connection here
		go s.handleClient(conn)
	}

}

func (s *SMTPServer) handleClient(conn net.Conn) {
	defer conn.Close()

	redeader := bufio.NewReader(conn)
	session := &Session{} // 新しいセッションを作成

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

		// ドメイン名の元の大文字小文字を保持
		//originalParts := strings.Fields(command)

		if len(parts) >= 1 && (parts[0] == "HELO" || parts[0] == "EHLO") {
			handleHelo(conn, parts, session)
			fmt.Println("Session after HELO/EHLO:", session)
		} else if len(parts) >= 1 && parts[0] == "MAIL" && strings.HasPrefix(parts[1], "FROM:") {
			handleMailFrom(conn, parts, session)
			fmt.Println("Session after MAIL FROM:", session)

		} else if len(parts) >= 1 && parts[0] == "RCPT" && strings.HasPrefix(parts[1], "TO:") {
			handleMailRcpt(conn, parts, session)
			fmt.Println("Session after RCPT TO:", session)

		} else if len(parts) >= 1 && parts[0] == "DATA" {
			s.handleData(conn, redeader, session)
			fmt.Println("Session after DATA:", session)

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

func (s *SMTPServer) handleData(conn net.Conn, reader *bufio.Reader, session *Session) {
	sendResponse(conn, "354 End data with <CR><LF>.<CR><LF>\r\n")
	var body strings.Builder
	for {
		line, err := reader.ReadString('\n')
		//fmt.Println("[", line, "]")
		fmt.Println(line)
		if err != nil {
			fmt.Println("Error reading data:", err)
			return
		}
		if strings.TrimSpace(line) == "." {
			fmt.Println("data end")
			break
		}
		body.WriteString(line)
	}
	session.Body = body.String()
	sendResponse(conn, "250 OK\r\n")

	// セッションに保持しているメールの情報を永続化させる
	// mailbox/userとしてディレクトリを作り、ファイルに保存する
	s.saveMail(session)

}
func (s *SMTPServer) saveMail(session *Session) {
	//メールの内容を構築
	mailContent := fmt.Sprintf("From: %s\r\nTo: %s\r\n\r\n%s",
		session.SenderMail,
		strings.Join(session.RecipientMail, ", "),
		session.Body,
	)

	//各受信者ごとにディレクトリを作成し、ファイルを保存
	for _, recipient := range session.RecipientMail {
		if s.IsSaveMail(recipient) == false {
			fmt.Println("Not local delivery. Skipping save for:", recipient)
			continue
		}

		// 受信者のメールアドレスからユーザー名を取得
		userName := strings.Split(recipient, "@")[0]
		dir := filepath.Join("mailbox", userName)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			continue
		}

		// ファイル名をタイムスタンプで作成
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s.mail", timestamp)
		filepath := filepath.Join(dir, filename)
		err = os.WriteFile(filepath, []byte(mailContent), 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			continue
		}

		fmt.Printf("Mail saved to %s\n", filepath)
	}
}

func handleRset(conn net.Conn) {
	sendResponse(conn, "250 OK\r\n")
}

func sendResponse(conn net.Conn, message string) {
	fmt.Println("send", message)
	conn.Write([]byte(message))
}
