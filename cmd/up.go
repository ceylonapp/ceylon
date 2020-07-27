package cmd

import (
	"ceylon/cli/mgt/virtualenv"
	"context"
	"github.com/spf13/cobra"
)

var up = &cobra.Command{
	Use: "up",
	Run: func(cmd *cobra.Command, args []string) {

		forceCreate, err := cmd.Flags().GetBool("forceCreate")
		if err != nil {
			panic(err)
		}
		deployManager := virtualenv.CreateInstance(context.Background())

		err = deployManager.Create(&virtualenv.CreateSettings{ForceCreate: forceCreate})
		err = deployManager.Prepare()
		err = deployManager.Run()
		//

		if err != nil {
			panic(err)
		}
	},
}
