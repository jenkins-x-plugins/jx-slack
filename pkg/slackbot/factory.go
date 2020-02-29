package slackbot

import (
	"fmt"

	"github.com/nlopes/slack"

	"k8s.io/client-go/kubernetes"

	slackapp "github.com/jenkins-x-labs/app-slack/pkg/apis/slack/v1alpha1"
	v1client "github.com/jenkins-x-labs/app-slack/pkg/client/clientset/versioned"
	jenkinsv1client "github.com/jenkins-x/jx/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx/pkg/jx/cmd"
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
	CommonOptions *cmd.CommonOptions
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

	commonOptions := cmd.NewCommonOptions(ns, factory)

	return &Clients{
		SlackAppClient: slackAppClient,
		Namespace:      ns,
		KubeClient:     kubeClient,
		JXClient:       jxClient,
		Factory:        factory,
		CommonOptions:  &commonOptions,
	}, nil
}

func CreateSlackBots(hmacSecretName string, port int) (*SlackBots, error) {
	if hmacSecretName == "" {
		hmacSecretName = DefaultHmacSecretName
	}
	if port == 0 {
		port = DefaultPort
	}
	c, err := CreateClients()
	if err != nil {
		return nil, err
	}
	// TODO make this a watch
	slackBots, err := c.SlackAppClient.SlackV1alpha1().SlackBots(c.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	optionsList := make([]*SlackBotOptions, 0)

	for _, slackBot := range slackBots.Items {
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
		options := SlackBotOptions{
			Clients:     c,
			Namespace:   watchNs,
			SlackClient: slackClient,

			PullRequests: slackBot.Spec.PullRequests,
			Pipelines:    slackBot.Spec.Pipelines,
			Statuses:     slackBot.Spec.Statuses,
			Timestamps:   make(map[string]map[string]*MessageReference, 0),
			SlackUserResolver: &SlackUserResolver{
				JXClient:    c.JXClient,
				Namespace:   c.Namespace,
				SlackClient: slackClient,
			},
		}
		optionsList = append(optionsList, &options)

	}
	return &SlackBots{
		Clients:        c,
		HmacSecretName: hmacSecretName,
		Port:           port,
		Items:          optionsList,
	}, nil
}

func (s *SlackBots) GetWebHookToken() ([]byte, error) {
	if s.HmacSecretName == "" || s.HmacSecretName == "REPLACE_ME" {
		// Not configured
		return nil, nil
	}
	secret, err := s.KubeClient.CoreV1().Secrets(s.Namespace).Get(s.HmacSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secret.Data["hmac"], nil
}
