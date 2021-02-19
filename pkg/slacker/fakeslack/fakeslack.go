package fakeslack

import (
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// FakeSlack the fake slack
type FakeSlack struct {
	UsersByEmail map[string]*slack.User
	Messages     map[string][]FakeMessage
}

type FakeMessage struct {
	Channel   string
	Timestamp string
	Options   []slack.MsgOption
}

// NewFakeSlack creates a new fake slack
func NewFakeSlack() *FakeSlack {
	return &FakeSlack{}
}

func (f *FakeSlack) OpenConversation(params *slack.OpenConversationParameters) (*slack.Channel, bool, bool, error) {
	return nil, false, false, nil
}

func (f *FakeSlack) SendMessage(channel string, options ...slack.MsgOption) (string, string, string, error) {
	if f.Messages == nil {
		f.Messages = map[string][]FakeMessage{}
	}

	timestamp := time.Now().String()
	msg := FakeMessage{
		Channel:   channel,
		Timestamp: timestamp,
		Options:   options,
	}

	f.Messages[channel] = append(f.Messages[channel], msg)
	return channel, timestamp, "", nil
}

func (f *FakeSlack) GetUserByEmail(email string) (*slack.User, error) {
	if f.UsersByEmail == nil {
		f.UsersByEmail = map[string]*slack.User{}
	}
	return f.UsersByEmail[email], nil
}

// AssertMessageCount asserts the message count for the given channel
func (f *FakeSlack) AssertMessageCount(t *testing.T, channel string, expectedCount int, message string) {
	if f.Messages == nil {
		f.Messages = map[string][]FakeMessage{}
	}
	messages := f.Messages[channel]
	require.Len(t, messages, expectedCount, "messages for channel %s for %s", channel, message)
}
