package handler

import (
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"log"
	"regexp"
	"strings"
	"wechat_test/ws"
)

func HandleMessage(dispatcher *openwechat.MessageMatchDispatcher,
	wsServer *ws.WebsocketServerMessageHandler) {

	// 处理群组消息并广播到 WebSocket 中
	dispatcher.OnGroup(func(ctx *openwechat.MessageContext) {
		// 跳过部分不支持的消息，通知消息、自己发送的消息
		skipSomeMessage(ctx)
		// 检测群名是否修改
		onGroupNameChange(ctx)

		groupSender, _ := ctx.SenderInGroup()
		self := groupSender.Self()
		list := self.MemberList
		marshal, _ := json.Marshal(list)
		wsServer.Broadcast(marshal)
		userName := groupSender.UserName
		displayName := groupSender.DisplayName
		nickName := groupSender.NickName
		avatarID := groupSender.AvatarID()
		message := ctx.Message.Content
		messageId := ctx.Message.NewMsgId
		toUserName := ctx.Message.ToUserName
		fromUserName := ctx.Message.FromUserName
		fmt.Println(userName, displayName, avatarID, nickName, message, messageId, toUserName, fromUserName)

		//拼接对象
		res := "userName=" + userName + ";displayName=" + displayName + ";avatarId=" + avatarID + ";nickName=" + nickName + ";message=" + message
		wsServer.Broadcast([]byte(res))
	})

}

/**跳过不支持的消息，通知消息、自己发送的消息*/
func skipSomeMessage(ctx *openwechat.MessageContext) {
	if ctx.IsNotify() || ctx.IsSendBySelf() {
		ctx.Abort()
	}
}

/**检测群名是否修改*/
func onGroupNameChange(msg *openwechat.MessageContext) {
	if !msg.IsSystem() {
		return
	}
	matches := regexp.MustCompile(`"(.*?)"修改群名为“(.*?)”`).FindAllString(msg.Content, -1)
	if len(matches) > 0 {
		parts := strings.SplitN(matches[0], "修改群名为", 2)
		userName := strings.Trim(parts[0], `"`)
		groupName := strings.TrimPrefix(strings.TrimSuffix(parts[1], `”`), `“`)
		log.Println("检测到群名片已修改", groupName, userName)
		_ = msg.AsRead()
		return
	}
}
