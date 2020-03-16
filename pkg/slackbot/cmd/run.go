package cmd

import (
	"time"

	"github.com/pkg/errors"

	"github.com/jenkins-x/jx/pkg/log"

	slackappapi "github.com/jenkins-x-labs/slack/pkg/apis/slack/v1alpha1"
	"github.com/jenkins-x-labs/slack/pkg/slackbot"
	jxcmd "github.com/jenkins-x/jx/pkg/cmd/helper"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

type SlackAppRunOptions struct {
	Cmd            *cobra.Command
	Args           []string
	HmacSecretName string
	Port           int
	clients        *slackbot.Clients
	Items          []*slackbot.SlackBotOptions
	botChannels    map[types.UID]chan struct{}
}

func NewCmdRun() *cobra.Command {
	var options = &SlackAppRunOptions{}

	var rootCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the jenkins-x App for Slack controller",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			jxcmd.CheckErr(err)
		},
	}
	rootCmd.Flags().StringVarP(&options.HmacSecretName, slackbot.DefaultHmacSecretName, "", "hmac-token",
		"The name of github webhook secret")
	rootCmd.Flags().IntVarP(&options.Port, "port", "p", slackbot.DefaultPort,
		"The port to run the prow external plugin server on")
	rootCmd.AddCommand(NewCmdHook())
	return rootCmd
}

func (o *SlackAppRunOptions) Run() error {
	var err error
	o.clients, err = slackbot.CreateClients()
	if err != nil {
		return err
	}

	o.botChannels = make(map[types.UID]chan struct{})

	slackBots := &slackappapi.SlackBot{}
	_, controller := cache.NewInformer(
		cache.NewListWatchFromClient(o.clients.SlackAppClient.SlackV1alpha1().RESTClient(), "slackbots", o.clients.Namespace,
			fields.Everything()),
		slackBots,
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				o.add(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				o.delete(oldObj)
				o.add(newObj)
			},
			DeleteFunc: func(obj interface{}) {
				o.delete(obj)
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)

	bots := slackbot.SlackBots{
		Clients:        o.clients,
		HmacSecretName: o.HmacSecretName,
		Port:           o.Port,
	}
	err = bots.ProwExternalPluginServer()
	if err != nil {
		return errors.Wrap(err, "failed to start prow plugin server")
	}
	return nil
}

func (o *SlackAppRunOptions) add(obj interface{}) {
	slackBot, ok := obj.(*slackappapi.SlackBot)
	if !ok {
		log.Logger().Infof("Object is not a PipelineActivity %#v\n", obj)
		return
	}

	bot, err := slackbot.CreateSlackBot(o.clients, slackBot)
	if err != nil {
		log.Logger().Warnf("failed to create slack bot for %s", slackBot.Name)
	}

	o.Items = append(o.Items, bot)

	// store the channel so we can update or delete it later if the SlackBot resource gets updated in the cluster
	o.botChannels[slackBot.UID] = bot.WatchActivities()

}

func (o *SlackAppRunOptions) delete(obj interface{}) {
	slackBot, ok := obj.(*slackappapi.SlackBot)
	if !ok {
		log.Logger().Infof("Object is not a PipelineActivity %#v\n", obj)
		return
	}
	if o.botChannels[slackBot.UID] != nil {
		close(o.botChannels[slackBot.UID])
		log.Logger().Info("SlackBot channel closed successfully")
		delete(o.botChannels, slackBot.UID)
		log.Logger().Infof("SlackBot %s deleted", slackBot.Name)
	} else {
		log.Logger().Warnf("No SlackBot named %s found so not deleted", slackBot.Name)
	}
}
