package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

	// コンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドラー
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	// ルーティングテーブルを作成
	myRoutes := map[string]string{
		"domain.com": "local",
		"test.com":   "localhost:2026",
	}
	// SMTPサーバーを作成して起動
	server := NewSMTPServer(port, myRoutes)
	server.Start(ctx)

}
