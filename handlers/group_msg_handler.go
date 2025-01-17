package handlers

import (
	"log"
	"strings"

	"github.com/869413421/wechatbot/config"
	"github.com/869413421/wechatbot/gtp"
	"github.com/eatmoreapple/openwechat"
)

var _ MessageHandlerInterface = (*GroupMessageHandler)(nil)

// GroupMessageHandler 群消息处理
type GroupMessageHandler struct {
}

// handle 处理消息
func (g *GroupMessageHandler) handle(msg *openwechat.Message) error {

	if msg.IsText() {
		if config.LoadConfig().AtActiveSwitch && msg.IsAt() {
			return g.ReplyText(msg)
		}
		if config.LoadConfig().ActiveGroupSwitch && strings.Contains(msg.Content, config.LoadConfig().ActiveKeyword) {
			return g.ReplyText(msg)
		}
	}
	return nil
}

// NewGroupMessageHandler 创建群消息处理器
func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{}
}

// ReplyText 发送文本消息到群
func (g *GroupMessageHandler) ReplyText(msg *openwechat.Message) error {
	// 接收群消息
	sender, err := msg.Sender()
	if err != nil {
		return err
	}
	group := openwechat.Group{User: sender}
	log.Printf("Received Group %v Text Msg : %v", group.NickName, msg.Content)

	// 替换掉@文本，然后向GPT发起请求
	replaceText := "@" + sender.Self.NickName
	requestText := strings.TrimSpace(strings.ReplaceAll(msg.Content, replaceText, ""))
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

	//回复自己在群里发的消息
	self, err := msg.Bot.GetCurrentUser()
	if err != nil {
		return err
	}
	if msg.IsSendBySelf() {
		friend := openwechat.Friend{User: &openwechat.User{UserName: msg.ToUserName}}
		self.SendTextToFriend(&friend, reply)
		return nil
	}

	// 获取@我的用户
	// groupSender, err := msg.SenderInGroup()
	// if err != nil {
	// 	log.Printf("get sender in group error :%v \n", err)
	// 	return err
	// }

	// 回复@我的用户
	// atText := "@" + groupSender.NickName + " "
	// replyText := atText + reply
	_, err = msg.ReplyText(reply)
	if err != nil {
		log.Printf("response group error: %v \n", err)
	}
	return err
}
