package cmd

import (
	"ceylon/cli/mgt"
	"context"
	"github.com/spf13/cobra"
)

var deploy = &cobra.Command{
	Use: "deploy",
	Run: func(cmd *cobra.Command, args []string) {

		deployManager := mgt.DeployManager{
			Context: context.Background(),
		}

		err := deployManager.Deploy()
		if err != nil {
			panic(err)
		}
	},
}
