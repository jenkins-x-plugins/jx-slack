package cmd

import (
	jxcmd "github.com/jenkins-x/jx/pkg/cmd/helper"
	"github.com/spf13/cobra"
)

type SlackAppHookOptions struct {
	Cmd  *cobra.Command
	Args []string
}

func NewCmdHook() *cobra.Command {
	var options = &SlackAppHookOptions{}

	var rootCmd = &cobra.Command{
		Use:   "hook",
		Short: "Hooks used by the app during install, upgrade and uninstall",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			jxcmd.CheckErr(err)
		},
	}
	rootCmd.AddCommand(NewCmdInstall())
	return rootCmd
}

func (o *SlackAppHookOptions) Run() error {
	return o.Cmd.Help()
}
