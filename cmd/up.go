package cmd

import (
	"ceylon/cli/mgt/docker"
	"context"
	"github.com/spf13/cobra"
)

var up = &cobra.Command{
	Use: "up",
	Run: func(cmd *cobra.Command, args []string) {

		forceCreate, err := cmd.Flags().GetBool("forceCreate")

		deployManager := docker.DeployManager{
			Context: context.Background(),
		}

		err = deployManager.Deploy(forceCreate)
		if err != nil {
			panic(err)
		}
	},
}
