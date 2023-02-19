package handlers

import (
	"log"
	"strings"

	"github.com/869413421/wechatbot/config"
	"github.com/869413421/wechatbot/gtp"
	"github.com/eatmoreapple/openwechat"
)

var _ MessageHandlerInterface = (*UserMessageHandler)(nil)

// UserMessageHandler 私聊消息处理
type UserMessageHandler struct {
}

// handle 处理消息
func (g *UserMessageHandler) handle(msg *openwechat.Message) error {
	if msg.IsText() {
		if config.LoadConfig().ActiveUserSwitch {
			if strings.Contains(msg.Content, config.LoadConfig().ActiveKeyword) {
				return g.ReplyText(msg)
			}
		} else {
			return g.ReplyText(msg)
		}

	}
	return nil
}

// NewUserMessageHandler 创建私聊处理器
func NewUserMessageHandler() MessageHandlerInterface {
	return &UserMessageHandler{}
}

// ReplyText 发送文本消息到群
func (g *UserMessageHandler) ReplyText(msg *openwechat.Message) error {
	// 接收私聊消息
	sender, err := msg.Sender()
	if err != nil {
		return err
	}
	log.Printf("Received User %v Text Msg : %v", sender.NickName, msg.Content)

	// 向GPT发起请求
	requestText := strings.TrimSpace(msg.Content)
	requestText = strings.Trim(requestText, "\n")
	requestText = strings.TrimSpace(strings.ReplaceAll(requestText, config.LoadConfig().ActiveKeyword, ""))
	reply, err := gtp.Completions(requestText)
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		msg.ReplyText(config.LoadConfig().ErrReplyWord)
		return err
	}
	if reply == "" {
		return nil
	}
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")

	self, err := msg.Bot.GetCurrentUser()
	if err != nil {
		return err
	}
	if msg.IsSendBySelf() {
		friend := openwechat.Friend{User: &openwechat.User{UserName: msg.ToUserName}}
		self.SendTextToFriend(&friend, reply)
		return nil
	}

	// 回复用户

	_, err = msg.ReplyText(reply)
	if err != nil {
		log.Printf("response user error: %v \n", err)
	}
	return err
}
