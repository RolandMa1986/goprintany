/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"goprintany/model"
	"goprintany/winspool"

	"github.com/spf13/cobra"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "inspect a printer",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		printers, err := winspool.GetPrinters()
		if err != nil {
			fmt.Println(err)
			return
		}
		var printer *model.Printer
		for _, p := range printers {
			if p.Name == printerName {
				printer = &p
				break
			}
		}

		if printer == nil {
			fmt.Printf("printer %s not found", printerName)
			return
		}
		body, err := json.MarshalIndent(*printer, "", "   ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(body))
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	inspectCmd.Flags().StringVarP(&printerName, "printer", "p", "", "Printer Name")
	printCmd.MarkFlagRequired("printer")
}
