package cmd

import (
	"ceylon/cli/mgt"
	"context"
	"github.com/spf13/cobra"
)

var up = &cobra.Command{
	Use: "up",
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
