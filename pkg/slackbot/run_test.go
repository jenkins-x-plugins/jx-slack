package slackbot

import (
	"testing"
)

func TestCreateSlackBot(t *testing.T) {
	/* TODO
	secretName := "test_secret"
	testToken := "123abc"
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			"token": []byte(testToken),
		},
	}
	fakeclient := fake.NewSimpleClientset(secret)

	tests := []struct {
		name     string
		slackBot *slackapp.SlackBot
		want     *slack.Client
		wantErr  bool
	}{
		{name: "secret_does_exist", slackBot: getSlackBot(secretName), want: clients.slackClientHelper.getSlackClient(testToken), wantErr: false},
		{name: "secret_does_not_exist", slackBot: getSlackBot("does_not_exist"), want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateSlackBot(clients, tt.slackBot)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSlackBot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("CreateSlackBot() got = nil")
			}
			if !tt.wantErr && !reflect.DeepEqual(got.SlackClient, tt.want) {
				t.Errorf("CreateSlackBot() got = %v, want %v", got, tt.want)
			}
		})
	}
	*/
}

/*
func getSlackBot(secretName string) *slackappapi.SlackBot {
	return &slackappapi.SlackBot{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test_slack_bot",
		},
		Spec: slackappapi.SlackBotSpec{
			TokenReference: slackappapi.ResourceReference{
				Kind: "Secret",
				Name: secretName,
			},
		},
	}
}

type fakeSlackClient struct {
	*slack.Client
}

func (f *fakeSlackClient) getSlackClient(token string, options ...slack.Option) *slack.Client {
	once.Do(startServer)
	return slack.New(token, slack.OptionAPIURL("http://"+serverAddr+"/"))
}


*/
