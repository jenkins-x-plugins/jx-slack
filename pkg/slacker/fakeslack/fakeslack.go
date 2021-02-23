package fakeslack

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jenkins-x/jx-helpers/v3/pkg/files"
	"github.com/jenkins-x/jx-helpers/v3/pkg/testhelpers"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

// FakeSlack the fake slack
type FakeSlack struct {
	UsersByEmail map[string]*slack.User
	Messages     map[string][]Message
}

type Message struct {
	Channel   string
	Timestamp string
	Options   []slack.MsgOption
}

type Attachment struct {
	Color      string   `json:"color,omitempty"`
	Fallback   string   `json:"fallback,omitempty"`
	CallbackID string   `json:"callback_id,omitempty"`
	Title      string   `json:"title,omitempty"`
	Actions    []Action `json:"actions,omitempty"`
	Timestamp  int      `json:"ts,omitempty"`
}

type Action struct {
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// NewFakeSlack creates a new fake slack
func NewFakeSlack() *FakeSlack {
	return &FakeSlack{}
}

func (f *FakeSlack) OpenConversation(_ *slack.OpenConversationParameters) (*slack.Channel, bool, bool, error) {
	return nil, false, false, nil
}

func (f *FakeSlack) SendMessage(channel string, options ...slack.MsgOption) (string, string, string, error) {
	if f.Messages == nil {
		f.Messages = map[string][]Message{}
	}

	timestamp := time.Now().String()
	msg := Message{
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
func (f *FakeSlack) AssertMessageCount(t *testing.T, channel string, expectedCount int, expectedMessageDir string, expectedMessagePrefix string, generateTestOutput bool, message string) []Attachment {
	if f.Messages == nil {
		f.Messages = map[string][]Message{}
	}
	messages := f.Messages[channel]
	require.Len(t, messages, expectedCount, "messages for channel %s for %s", channel, message)

	dir := expectedMessageDir
	var err error
	if generateTestOutput {
		err = os.MkdirAll(dir, files.DefaultDirWritePermissions)
		require.NoError(t, err, "failed to create dir %s", dir)
	} else {
		dir, err = ioutil.TempDir("", "")
		require.NoError(t, err, "failed to create a temp dir")
	}
	var attachments []Attachment
	for i := 0; i < expectedCount; i++ {
		message := messages[i]
		_, values, err := slack.UnsafeApplyMsgOptions("fakeToken", channel, "fakeapiurl", message.Options...)
		require.NoError(t, err, "failed to render message %d for %s", i, message)
		attachmentsJSON := values.Get("attachments")
		require.NotEmpty(t, attachmentsJSON, "no attachments JSON found for message %d of %s", i, message)

		//t.Logf("got message: %s", attachmentsJSON)

		err = json.Unmarshal([]byte(attachmentsJSON), &attachments)
		require.NoError(t, err, "failed to parse attachments JSON %s for message %d of %s", attachmentsJSON, i, message)

		// lets clear the timestamps
		for i := range attachments {
			a := &attachments[i]
			a.Timestamp = 0
		}

		out := &strings.Builder{}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		err = enc.Encode(attachments)
		require.NoError(t, err, "failed to marshal attachments to JSON for message %d of %s", i, message)

		attachmentsJSON = out.String()

		fileName := expectedMessagePrefix + "-" + strconv.Itoa(i+1) + ".json"
		path := filepath.Join(dir, fileName)

		err = ioutil.WriteFile(path, []byte(attachmentsJSON), files.DefaultFileWritePermissions)
		require.NoError(t, err, "failed to save file %s", path)

		if generateTestOutput {
			t.Logf("generated test output %s\n", path)
		} else {
			testhelpers.AssertEqualFileText(t, filepath.Join(expectedMessageDir, fileName), path)
		}
	}
	return attachments
}
