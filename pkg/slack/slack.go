package slack

import (
	"log"

	"github.com/slack-go/slack"
)

type MessageInfo struct {
	Api       slack.Client
	ChannelID string
	UserName  string
	IconEmoji string
	Message   string
}

func NewSlackMessageInfo(token, channelId, message string) *MessageInfo {
	return &MessageInfo{
		Api:       *slack.New(token),
		ChannelID: channelId,
		Message:   message,
	}
}

func (i *MessageInfo) PostMessage() error {
	if _, _, err := i.Api.PostMessage(
		i.ChannelID,
		slack.MsgOptionText(i.Message, false),
	); err != nil {
		return err
	}
	log.Printf("[INFO] Post message %v", i)
	return nil
}
