package main

import (
	"context"
	"github.com/eatmoreapple/openwechat"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
	"wechat_test/handler"
	"wechat_test/login"
	"wechat_test/ws"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// 创建一个新的微信机器人
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// 尝试热登录
	login.HandleHotLogin(bot)

	// 获取回调函数
	dispatcher := openwechat.NewMessageMatchDispatcher()
	dispatcher.SetAsync(true)

	// 处理外部上报过来的指令消息
	onMessageFn := func(message []byte) {
		slog.Info("收到消息", "msg", string(message))
	}

	// 设置消息转发,websocket服务器方式
	wsServer := ws.NewWebsocketServerMessageHandler(ctx, time.Second*10, onMessageFn)

	bot.MessageHandler = dispatcher.AsMessageHandler()

	//启动HTTP服务器
	go func() {
		log.Println("Listening on :8080...")
		err := http.ListenAndServe(":8080", wsServer)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	// 处理回调函数
	handler.HandleMessage(dispatcher, wsServer)

	// 阻塞以保持程序运行
	bot.Block()
}
