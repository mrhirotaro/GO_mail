package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type SMTPClient struct {
	// クライアントの状態を管理するフィールドを追加できます
	serverAddress string
}

func NewSMTPClient(serverAddress string) *SMTPClient {
	return &SMTPClient{serverAddress: serverAddress}
}

func (c *SMTPClient) Start(domain, senderMail, recipientMail, body, data string) error {
	// サーバーへの接続処理を実装します
	fmt.Println("Connecting to server at", c.serverAddress)

	// 1 サーバーに接続
	conn, err := net.Dial("tcp", c.serverAddress)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return err
	}
	defer conn.Close()
	fmt.Println("Connected to server")

	// 2 buifoio.Readerを作成
	readeer := bufio.NewReader(conn)

	// 3 gritingメッセージを受信
	readGritingMessage(readeer)

	// 4 HELOコマンドを送信
	sendHeloCommand(conn, readeer, domain)

	// 5 MAIL FROMコマンドを送信
	sendMailFromCommand(conn, readeer, senderMail)

	// 6 RCPT TOコマンドを送信
	sendRcptToCommand(conn, readeer, recipientMail)
	// 7 DATAコマンドを送信
	sendDataCommand(conn, readeer, data)

	// 8 メール本文を送信
	sendBody(conn, readeer, body)

	// 9 QUITコマンドを送信
	sendQuitCommand(conn, readeer)

	return nil
}

func readGritingMessage(reader *bufio.Reader) {
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading greeting message:", err)
		return
	}
	fmt.Println("Server greeting: ", line)
}

func sendHeloCommand(conn net.Conn, reader *bufio.Reader, domain string) {
	response, err := sendCommand(conn, reader, "HELO "+domain)
	if err != nil {
		fmt.Println("Error sending HELO command:", err)
		return
	}
	fmt.Println("HELO response: ", response)
	if response != "250" {
		fmt.Println("HELO command failed with response:", response)
	}
}

func sendMailFromCommand(conn net.Conn, reader *bufio.Reader, senderMail string) {
	response, err := sendCommand(conn, reader, "MAIL FROM:<"+senderMail+">")
	if err != nil {
		fmt.Println("Error sending MAIL FROM command:", err)
		return
	}
	fmt.Println("MAIL FROM response: ", response)
	if response != "250" {
		fmt.Println("MAIL FROM command failed with response:", response)
	}
}

func sendRcptToCommand(conn net.Conn, reader *bufio.Reader, recipientMail string) {
	response, err := sendCommand(conn, reader, "RCPT TO:<"+recipientMail+">")
	if err != nil {
		fmt.Println("Error sending RCPT TO command:", err)
		return
	}
	fmt.Println("RCPT TO response: ", response)
	if response != "250" {
		fmt.Println("RCPT TO command failed with response:", response)
	}
}

func sendDataCommand(conn net.Conn, reader *bufio.Reader, data string) {
	response, err := sendCommand(conn, reader, data)
	if err != nil {
		fmt.Println("Error sending DATA command:", err)
		return
	}
	fmt.Println("DATA response: ", response)
	if response != "354" {
		fmt.Println("DATA command failed with response:", response)
	}
}

func sendBody(conn net.Conn, reader *bufio.Reader, body string) {
	// メール本文の送信
	response, _ := sendCommand(conn, reader, body+"\r\n.")
	// 終了のドットを送信
	//end, _ := sendCommand(conn, reader, ".")
	fmt.Println("Body response: ", response)
	//fmt.Println("End response: ", end)
}

func sendQuitCommand(conn net.Conn, reader *bufio.Reader) {
	response, err := sendCommand(conn, reader, "QUIT")
	if err != nil {
		fmt.Println("Error sending QUIT command:", err)
		return
	}
	fmt.Println("QUIT response: ", response)
	if response != "221" {
		fmt.Println("QUIT command failed with response:", response)
	}
}

func sendCommand(conn net.Conn, reader *bufio.Reader, command string) (string, error) {
	_, err := conn.Write([]byte(command + "\r\n"))
	if err != nil {
		fmt.Println("Error sending command:", err)
		return "", err
	}
	fmt.Println("Sent command:", command)

	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}
	fmt.Println("Server response: ", response)
	parts := strings.Fields(response)
	return parts[0], nil
}
