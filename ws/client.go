package ws

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
	"time"
)

type Client interface {
	Serve(ctx context.Context) error
	Close()
	SendMessage(message []byte) error
}
type onMessage func(message []byte)

type client struct {
	*websocket.Conn
	heartbeat         time.Duration
	messageBufferPool chan []byte
	exit              chan error
	cancelFn          context.CancelFunc
	onMessage         onMessage
}

func newClient(conn *websocket.Conn, heartbeat time.Duration, onMessage onMessage) Client {
	conn.SetReadLimit(1024 * 1024 * 10)
	return &client{
		Conn:              conn,
		heartbeat:         heartbeat,
		messageBufferPool: make(chan []byte, 5),
		exit:              make(chan error),
		onMessage:         onMessage,
	}
}

func (c *client) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFn = cancel
	var ticker *time.Ticker
	if c.heartbeat > 0 {
		ticker = time.NewTicker(c.heartbeat)
	}
	defer func() {
		cancel()
		if ticker != nil {
			ticker.Stop()
		}
		close(c.messageBufferPool)
		_ = c.Conn.Close()
	}()
	go c.receiveMessage()
	go c.sendMessage()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := c.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(2*time.Second)); err != nil {
				slog.Error("心跳消息出错", "error", err)
				return err
			}
		case err := <-c.exit:
			return err
		}
	}
}
func (c *client) Close() {
	if c.cancelFn != nil {
		c.cancelFn()
	} else {
		_ = c.Conn.Close()
	}
}
func (c *client) SendMessage(message []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	c.messageBufferPool <- message
	return nil
}

func (c *client) receiveMessage() {
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			slog.Error("读取消息出错", "error", err)
			c.exit <- err
			return
		}
		switch messageType {
		case websocket.PingMessage:
			// 当接收到 Ping 消息时，自动发送 Pong 消息
			if err := c.WriteMessage(websocket.PongMessage, message); err != nil {
				slog.Error("pong消息出错", "error", err)
				c.exit <- err
				return
			}
		default:
			if c.onMessage != nil {
				c.onMessage(message)
			} else {
				slog.Info("收到消息", "messageType", messageType, "message", string(message))
			}
		}
	}
}

// redirectMessage 将缓冲队列中的信息转发到服务器
func (c *client) sendMessage() {
	for {
		message := <-c.messageBufferPool
		err := c.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			slog.Error("发送消息出错", "error", err)
			c.exit <- err
			return
		}
	}
}
