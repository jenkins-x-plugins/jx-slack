package cmd

import (
	"net/http"
	"strconv"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/pkg/errors"

	"github.com/jenkins-x/jx-logging/pkg/log"

	jxcmd "github.com/jenkins-x/jx/v2/pkg/cmd/helper"
	slackappapi "github.com/jenkins-x/slack/pkg/apis/slack/v1alpha1"
	informers "github.com/jenkins-x/slack/pkg/client/informers/externalversions"
	"github.com/jenkins-x/slack/pkg/slackbot"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

type SlackAppRunOptions struct {
	Cmd            *cobra.Command
	Args           []string
	HmacSecretName string
	Port           int
	clients        *slackbot.GlobalClients
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

	log.Logger().Infof("Watching slackbots in namespace %s\n", o.clients.Namespace)

	factory := informers.NewSharedInformerFactoryWithOptions(o.clients.SlackAppClient, 0, informers.WithNamespace(o.clients.Namespace))

	informer := factory.Slack().V1alpha1().SlackBots().Informer()

	stopper := make(chan struct{})
	defer close(stopper)

	defer runtime.HandleCrash()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.add,
		UpdateFunc: o.onUpdate,
		DeleteFunc: o.delete,
	})

	go informer.Run(stopper)

	isLighthouse := false
	_, err = o.clients.KubeClient.AppsV1().Deployments(o.clients.Namespace).Get("tide", metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			isLighthouse = true
		} else {
			return err
		}
	}

	bots := slackbot.SlackBots{
		GlobalClients:  o.clients,
		HmacSecretName: o.HmacSecretName,
		Port:           o.Port,
		IsLighthouse:   isLighthouse,
	}
	handler := bots.ExternalPluginServer()
	err = http.ListenAndServe("0.0.0.0:"+strconv.Itoa(o.Port), handler)
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
}

func (o SlackAppRunOptions) onUpdate(oldObj interface{}, newObj interface{}) {
	o.delete(newObj)
	o.add(newObj)
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
