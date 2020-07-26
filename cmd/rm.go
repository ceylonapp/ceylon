package cmd

import (
	"ceylon/cli/mgt/docker"
	"context"
	"github.com/spf13/cobra"
)

var rm = &cobra.Command{
	Use: "rm",
	Run: func(cmd *cobra.Command, args []string) {

		isPrune, err := cmd.Flags().GetBool("prune")

		deployManager := docker.DeployManager{
			Context: context.Background(),
		}

		err = deployManager.Destroy(isPrune)
		if err != nil {
			panic(err)
		}
	}}
