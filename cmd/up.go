package cmd

import (
	"ceylon/cli/mgt"
	"context"
	"github.com/spf13/cobra"
)

var up = &cobra.Command{
	Use: "up",
	Run: func(cmd *cobra.Command, args []string) {

		forceCreate, err := cmd.Flags().GetBool("forceCreate")

		deployManager := mgt.DeployManager{
			Context: context.Background(),
		}

		err = deployManager.Deploy(forceCreate)
		if err != nil {
			panic(err)
		}
	},
}
