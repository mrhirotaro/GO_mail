package main

//メールアドレスをもらったらドメイン名を切り出す

import (
	"strings"
)

func (s *SMTPServer) IsSaveMail(email string) bool {
	domain := extractDomain(email)
	dest := s.DomainRuts[domain]
	if dest == "local" {
		return true
	}
	return false
}

func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
