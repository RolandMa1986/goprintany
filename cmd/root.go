/*
Copyright © 2023 Roland Ma

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "print anything",
	Short: "A cli tool for windows printer",
	Long:  `A cli tool for windows printer, print anything from cli`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
