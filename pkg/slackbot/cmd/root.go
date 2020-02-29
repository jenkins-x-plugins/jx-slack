package cmd

import (
	jxcmd "github.com/jenkins-x/jx/pkg/jx/cmd"
	"github.com/spf13/cobra"
)

type SlackAppOptions struct {
	Cmd  *cobra.Command
	Args []string
}

func NewCmdRoot() *cobra.Command {
	var options = &SlackAppOptions{}

	var rootCmd = &cobra.Command{
		Use:   "app-slack",
		Short: "The jenkins-x App for Slack allows you to reports pipelines and review requests in Slack",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			jxcmd.CheckErr(err)
		},
	}
	rootCmd.AddCommand(NewCmdHook())
	rootCmd.AddCommand(NewCmdRun())
	return rootCmd
}

func (o *SlackAppOptions) Run() error {
	return o.Cmd.Help()
}
