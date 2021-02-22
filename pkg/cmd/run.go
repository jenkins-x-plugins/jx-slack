package cmd

import (
	"context"
	"github.com/jenkins-x-plugins/jx-slack/pkg/slackbot"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	var o = &slackbot.SlackBotOptions{}

	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Run the Jenkins X slack bot",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}

	err := envconfig.Process(context.TODO(), &o.SlackOptions)
	if err != nil {
		log.Logger().Warnf("failed to process environment variables: %s", err.Error())
	}

	cmd.Flags().StringVarP(&o.Dir, "dir", "d", o.Dir, "the directory to point to a git clone of your development repository. Mostly used for development and testing")
	cmd.Flags().StringVarP(&o.GitURL, "git-url", "u", o.GitURL, "the git URL to clone for the dev cluster git repository")
	cmd.Flags().StringVarP(&o.SlackToken, "slack-token", "t", o.SlackToken, "the slack token")
	return cmd
}
