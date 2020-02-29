package cmd

import (
	"github.com/jenkins-x-labs/app-slack/pkg/slackbot"
	jxcmd "github.com/jenkins-x/jx/pkg/jx/cmd"
	"github.com/spf13/cobra"
)

type SlackAppRunOptions struct {
	Cmd            *cobra.Command
	Args           []string
	HmacSecretName string
	Port           int
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

	bots, err := slackbot.CreateSlackBots(o.HmacSecretName, o.Port)
	if err != nil {
		return err
	}
	err = bots.Run()
	if err != nil {
		return err
	}
	return bots.ProwExternalPluginServer()
}
