/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"goprintany/winspool"

	"github.com/spf13/cobra"

	"github.com/cheynewallace/tabby"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all printers",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		t := tabby.New()
		t.AddHeader("Name", "Location", "Driver", "status")
		printers, err := winspool.GetPrinters()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, p := range printers {
			t.AddLine(p.Name, p.Location, p.Model, p.State.State)
		}

		t.Print()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("short", "t", false, "Get printer's name only")
}
