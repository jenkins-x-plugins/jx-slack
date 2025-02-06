package slacker

import (
	"github.com/slack-go/slack"
)

// Interface the main slack interface we need
// which is a small subset of the slack API so its easier to fake
type Interface interface {
	OpenConversation(params *slack.OpenConversationParameters) (*slack.Channel, bool, bool, error)

	SendMessage(channel string, options ...slack.MsgOption) (string, string, string, error)

	GetUserByEmail(email string) (*slack.User, error)
}
