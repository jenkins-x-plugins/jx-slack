package slackbot

import (
	slackappapi "github.com/jenkins-x-labs/slack/pkg/apis/slack/v1alpha1"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxclient"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"

	"github.com/slack-go/slack"

	"k8s.io/client-go/kubernetes"

	slackapp "github.com/jenkins-x-labs/slack/pkg/apis/slack/v1alpha1"
	jenkinsv1client "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
)

const (
	DefaultHmacSecretName = "hmac-token"
	DefaultPort           = 8080
)

// SlackBotOptions contains options for the SlackBot
type SlackBotOptions struct {
	Namespace         string
	KubeClient        kubernetes.Interface
	JXClient          jenkinsv1client.Interface
	SlackClient       *slack.Client
	ScmClient         *scm.Client
	Name              string
	Pipelines         []slackapp.SlackBotMode
	PullRequests      []slackapp.SlackBotMode
	Statuses          slackapp.Statuses
	Orgs              []slackapp.Org
	Timestamps        map[string]map[string]*MessageReference
	SlackUserResolver SlackUserResolver

	HmacSecretName string
	Port           int
}

type SlackBots struct {
	HmacSecretName string
	Items          []*SlackBotOptions
	Port           int
}

// Validate configures the clients for the slack bot
func (o *SlackBotOptions) Validate(slackBot *slackapp.SlackBot) error {
	if o.SlackClient == nil {
		token := os.Getenv("SLACK_TOKEN")
		if token == "" {
			return errors.Errorf("no $SLACK_TOKEN defined")
		}

		u := os.Getenv("SLACK_URL")
		if u != "" {
			log.Logger().Infof("using slack URL %s", u)
			o.SlackClient = slack.New(token, slack.OptionAPIURL(u))
		} else {
			o.SlackClient = slack.New(token)
		}
	}
	if o.Name == "" {
		o.Name = slackBot.Name
	}
	o.Pipelines = slackBot.Spec.Pipelines
	o.PullRequests = slackBot.Spec.PullRequests
	o.Statuses = slackBot.Spec.Statuses

	var err error
	o.KubeClient, o.Namespace, err = kube.LazyCreateKubeClientAndNamespace(o.KubeClient, o.Namespace)
	if err != nil {
		return err
	}

	o.JXClient, err = jxclient.LazyCreateJXClient(o.JXClient)
	if err != nil {
		return err
	}

	if o.ScmClient == nil {
		o.ScmClient, err = factory.NewClientFromEnvironment()
		if err != nil {
			return errors.Wrapf(err, "failed to create SCM client")
		}
	}
	if slackBot.Spec.Namespace != "" {
		o.Namespace = slackBot.Spec.Namespace
	}
	o.SlackUserResolver = NewSlackUserResolver(o.SlackClient, o.JXClient, o.Namespace)
	return nil
}

func (o *SlackBotOptions) Run() error {
	defer runtime.HandleCrash()

	channel := "testing-bot"

	slackBot := &slackappapi.SlackBot{
		Spec: slackappapi.SlackBotSpec{
			PullRequests: []slackappapi.SlackBotMode{
				{
					DirectMessage:   true,
					NotifyReviewers: false,
					Channel:         channel,
					Orgs:            nil,
					IgnoreLabels:    nil,
				},
			},
			Pipelines: []slackappapi.SlackBotMode{
				{
					DirectMessage:   true,
					NotifyReviewers: false,
					Channel:         channel,
					Orgs:            nil,
					IgnoreLabels:    nil,
				},
			},
			Statuses: slackappapi.Statuses{},
		},
	}

	err := o.Validate(slackBot)
	if err != nil {
		return errors.Wrapf(err, "failed to validate options")
	}

	log.Logger().Infof("Watching slackbots in namespace %s\n", o.Namespace)

	// store the channel so we can update or delete it later if the SlackBot resource gets updated in the cluster
	//o.botChannels[slackBot.UID] = bot.WatchActivities()
	o.WatchActivities()
	return nil
}
