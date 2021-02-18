package cmd

import (
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/spf13/cobra"
)

type SlackAppOptions struct {
	Cmd  *cobra.Command
	Args []string
}

func NewCmdRoot() *cobra.Command {
	var options = &SlackAppOptions{}

	var rootCmd = &cobra.Command{
		Use:   "slack",
		Short: "The jenkins-x App for Slack allows you to reports pipelines and review requests in Slack",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			helper.CheckErr(err)
		},
	}
	rootCmd.AddCommand(NewCmdRun())
	return rootCmd
}

func (o *SlackAppOptions) Run() error {
	return o.Cmd.Help()
}
