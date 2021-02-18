package cmd

import (
	"github.com/jenkins-x-labs/slack/pkg/slackbot"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	var o = &slackbot.SlackBotOptions{}

	var rootCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the jenkins-x App for Slack controller",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	rootCmd.Flags().StringVarP(&o.HmacSecretName, slackbot.DefaultHmacSecretName, "", "hmac-token",
		"The name of github webhook secret")
	rootCmd.Flags().IntVarP(&o.Port, "port", "p", slackbot.DefaultPort,
		"The port to run the prow external plugin server on")
	return rootCmd
}
