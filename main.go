package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	// メイン関数はserver.goに移動した
	// TODO: 引数を受け取ってサーバーを起動するようにする

	port := 1025 // デフォルトのポート番号
	if len(os.Args) > 1 {
		// ポート番号を検証
		var err error
		port, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("Invalid port number. Using default port 1025.")
		}

	}

	fmt.Println("server on port:", port)

	// SMTPサーバーを作成して起動
	server := NewSMTPServer(port)
	server.Start()

}
