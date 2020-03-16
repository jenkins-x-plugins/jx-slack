package slackbot

import (
	"fmt"

	"github.com/nlopes/slack"

	"k8s.io/client-go/kubernetes"

	slackapp "github.com/jenkins-x-labs/slack/pkg/apis/slack/v1alpha1"
	v1client "github.com/jenkins-x-labs/slack/pkg/client/clientset/versioned"
	jenkinsv1client "github.com/jenkins-x/jx/pkg/client/clientset/versioned"
	cmd "github.com/jenkins-x/jx/pkg/cmd/clients"
	opts "github.com/jenkins-x/jx/pkg/cmd/opts"
	"github.com/jenkins-x/jx/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultHmacSecretName = "hmac-token"
	DefaultPort           = 8080
)

type Clients struct {
	SlackAppClient v1client.Interface
	Namespace      string
	KubeClient     kubernetes.Interface
	JXClient       jenkinsv1client.Interface
	Factory        cmd.Factory

	// TODO hacky and probably going to break a lot, but until Git Provider stuff is better unwound...
	CommonOptions *opts.CommonOptions
}

// SlackBotOptions contains options for the SlackBot
type SlackBotOptions struct {
	*Clients

	SlackClient       *slack.Client
	Pipelines         []slackapp.SlackBotMode
	PullRequests      []slackapp.SlackBotMode
	Namespace         string
	Statuses          slackapp.Statuses
	Orgs              []slackapp.Org
	Timestamps        map[string]map[string]*MessageReference
	SlackUserResolver *SlackUserResolver

	HmacSecretName string
	Port           int
}

type SlackBots struct {
	*Clients
	HmacSecretName string
	Items          []*SlackBotOptions
	Port           int
}

func createSlackAppClient(f cmd.Factory) (v1client.Interface, string, error) {
	config, err := f.CreateKubeConfig()
	kuber := kube.NewKubeConfig()
	if err != nil {
		return nil, "", err
	}
	kubeConfig, _, err := kuber.LoadConfig()
	if err != nil {
		return nil, "", err
	}
	ns := kube.CurrentNamespace(kubeConfig)
	client, err := v1client.NewForConfig(config)
	if err != nil {
		return nil, ns, err
	}
	return client, ns, err
}

func CreateClients() (*Clients, error) {
	factory := cmd.NewFactory()

	slackAppClient, ns, err := createSlackAppClient(factory)
	if err != nil {
		return nil, err
	}

	kubeClient, ns, err := factory.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	jxClient, ns, err := factory.CreateJXClient()
	if err != nil {
		return nil, err
	}

	commonOptions := opts.NewCommonOptionsWithFactory(factory)

	return &Clients{
		SlackAppClient: slackAppClient,
		Namespace:      ns,
		KubeClient:     kubeClient,
		JXClient:       jxClient,
		Factory:        factory,
		CommonOptions:  &commonOptions,
	}, nil
}

// CreateSlackBot configures a SlackBot
func CreateSlackBot(c *Clients, slackBot *slackapp.SlackBot) (*SlackBotOptions, error) {

	// Fetch the resource reference for the token
	if slackBot.Spec.TokenReference.Kind != "Secret" {
		return nil, fmt.Errorf("expected token of kind Secret but got %s for %s", slackBot.Spec.TokenReference.Kind,
			slackBot.Name)
	}
	secret, err := c.KubeClient.CoreV1().Secrets(c.Namespace).Get(slackBot.Spec.TokenReference.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	token, ok := secret.Data["token"]
	if !ok {
		return nil, fmt.Errorf("expected key token in field data")
	}
	watchNs := c.Namespace
	if slackBot.Spec.Namespace != "" {
		watchNs = slackBot.Spec.Namespace
	}

	slackClient := slack.New(string(token))
	return &SlackBotOptions{
		Clients:      c,
		SlackClient:  slackClient,
		Pipelines:    slackBot.Spec.Pipelines,
		PullRequests: slackBot.Spec.PullRequests,
		Namespace:    watchNs,
		Statuses:     slackBot.Spec.Statuses,
		Timestamps:   make(map[string]map[string]*MessageReference, 0),
		SlackUserResolver: &SlackUserResolver{
			JXClient:    c.JXClient,
			Namespace:   c.Namespace,
			SlackClient: slackClient,
		},
	}, nil
}
