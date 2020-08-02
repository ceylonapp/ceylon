package cmd

import (
	"ceylon/cli/mgt/virtualenv"
	"context"
	"github.com/spf13/cobra"
	"os"
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
		err = os.Chdir(deployManager.ProjectPath)
		if err != nil {
			panic(err)
		}
		err = deployManager.Run()
		//

		if err != nil {
			panic(err)
		}
	},
}
