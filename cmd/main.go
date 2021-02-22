package main

import (
	"fmt"
	"os"

	"github.com/jenkins-x-plugins/jx-slack/pkg/cmd"
)

func main() {
	rootCmd := cmd.NewCmdRoot()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
