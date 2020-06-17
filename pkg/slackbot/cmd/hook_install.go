package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jenkins-x-labs/slack/pkg/slackbot"
	"k8s.io/test-infra/prow/plugins"

	jxcmd "github.com/jenkins-x/jx/v2/pkg/cmd/helper"
	"github.com/jenkins-x/jx/v2/pkg/prow"
	"github.com/spf13/cobra"
)

type SlackAppInstallOptions struct {
	Cmd  *cobra.Command
	Args []string
}

const (
	serviceNameEnvVar = "SERVICE_NAME"
)

func NewCmdInstall() *cobra.Command {
	var options = &SlackAppInstallOptions{}

	var command = &cobra.Command{
		Use:   "install",
		Short: "Installs the hook for the jenkins-x-labs Core Slack App",
		Long: fmt.Sprintf(`
	The following environment variables are used:
	
	* %s - the name of the service on which the Prow External Plugin server is exposed 
`, serviceNameEnvVar),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			jxcmd.CheckErr(err)
		},
	}
	return command
}

func (o *SlackAppInstallOptions) Run() error {
	clients, err := slackbot.CreateClients()
	if err != nil {
		return err
	}
	serviceName := os.Getenv(serviceNameEnvVar)
	if !strings.HasPrefix(serviceName, "http") {
		serviceName = fmt.Sprintf("http://%s", serviceName)
	}
	return prow.AddExternalPlugins(clients.KubeClient, nil, clients.Namespace, plugins.ExternalPlugin{
		Name:     "slack",
		Endpoint: serviceName,
	})
}
