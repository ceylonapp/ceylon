package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "version",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		log.Println("version 1.0.0")
	},
}

func Execute() {
	var rootCmd = &cobra.Command{Use: "ceylon", Run: func(cmd *cobra.Command, args []string) {

	}}
	rootCmd.AddCommand(deploy)
	//rootCmd.AddCommand(cmdPrint, cmdEcho)
	//cmdEcho.AddCommand(cmdTimes)
	rootCmd.Execute()
}
