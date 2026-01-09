package main

//import "fmt"

func main() {
	//fmt.Println("Start client main")

	server_address := "localhost:1025"
	client := NewSMTPClient(server_address)
	domain := "inu.com"
	senderMail := "user1@" + domain
	recipientMail := "user2@" + domain
	data := "DATA"
	body := "Hello, this is a test email."
	client.Start(domain, senderMail, recipientMail, body, data)

}
