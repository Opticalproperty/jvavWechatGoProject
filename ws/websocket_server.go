package ws

import (
	"context"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"time"
)

type WebsocketServerMessageHandler struct {
	upgrader   websocket.Upgrader
	messages   chan []byte
	heartbeat  time.Duration
	register   chan Client
	unregister chan Client
	clients    map[Client]struct{}
	onMessage  onMessage
}

func NewWebsocketServerMessageHandler(ctx context.Context, heartbeat time.Duration, onMessage onMessage) *WebsocketServerMessageHandler {
	h := &WebsocketServerMessageHandler{
		heartbeat: heartbeat,
		messages:  make(chan []byte, 10), // 消息缓冲区
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		register:   make(chan Client),
		unregister: make(chan Client),
		clients:    make(map[Client]struct{}),
		onMessage:  onMessage,
	}
	// 最低5s心跳
	if h.heartbeat > 0 && h.heartbeat < time.Second*5 {
		h.heartbeat = time.Second * 5
	}
	go h.serve(ctx)
	return h
}

func (h *WebsocketServerMessageHandler) serve(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()
	go h.SendMessage()
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
			go func() {
				_ = c.Serve(ctx)
			}()
		case c := <-h.unregister:
			delete(h.clients, c)
			c.Close()
		case <-ctx.Done():
			close(h.register)
			close(h.unregister)
			close(h.messages)
			clear(h.clients)
			return
		}
	}
}

func (h *WebsocketServerMessageHandler) SendMessage() {
	for message := range h.messages {
		for c := range h.clients {
			if err := c.SendMessage(message); err != nil {
				slog.Error("发送消息失败", "err", err)
				h.unregister <- c
			}
		}
	}
}

// 新增 Broadcast 方法直接发送消息到所有 WebSocket 客户端
func (h *WebsocketServerMessageHandler) Broadcast(message []byte) {
	h.messages <- message
}

//func (h *WebsocketServerMessageHandler) Register(dispatcher *openwechat.MessageMatchDispatcher) {
//	dispatcher.OnText(func(ctx *openwechat.MessageContext) {
//		h.messages <- []byte(ctx.Message.Content)
//	})
//}

func (h *WebsocketServerMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws upgrade", "err", err)
	}
	h.register <- newClient(conn, h.heartbeat, h.onMessage)
}
