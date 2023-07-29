/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"goprintany/winspool"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	printerName string = ""
	filePath    string = ""
	printType   string = "RAW"
)

var (
	ppi        int32   = 96    // windows default ppi
	rotation   int32   = 0     //rotation angle in degrees,eg: 0, 90, 180, 270
	scale      float32 = 1.0   // scale factor of the image. 1.0 means 100%.
	fitToPaper bool    = false // fit to paper
)

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "print a file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Print: "+strings.Join(args, " "), printerName, filePath, printType)

		switch printType {
		case "RAW":
			err := printerRaw()
			if err != nil {
				fmt.Println(err)
			}
		case "IMAGE":
			err := printerImage()
			if err != nil {
				fmt.Println(err)
			}
		default:
			fmt.Println("Not support print type: ", printType)
		}
	},
}

func printerRaw() error {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	file_name := filepath.Base(filePath)

	jobID, err := winspool.RawDataToPrinter(printerName, file_name, buf)
	if err != nil {
		return err
	}
	fmt.Println("jobID: ", jobID)
	return nil
}

func printerImage() error {
	catFile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	imageData, _, err := image.Decode(catFile)
	if err != nil {
		return err
	}
	file_name := filepath.Base(filePath)
	jobID, err := winspool.ImageDataToPrinter(printerName, file_name, imageData, rotation, ppi, fitToPaper)
	if err != nil {
		return err
	}

	fmt.Println("jobID: ", jobID)

	return nil
}

func init() {
	rootCmd.AddCommand(printCmd)
	printCmd.Flags().StringVarP(&printType, "type", "t", "IMAGE", "Data type send to printer：RAW，IMAGE，PDF")
	printCmd.Flags().StringVarP(&printerName, "printer", "p", "", "Printer Name")
	printCmd.Flags().StringVarP(&filePath, "file", "f", "", "file path")
	printCmd.Flags().Int32VarP(&ppi, "ppi", "i", 96, "Pixels per inch of the image")
	printCmd.Flags().Int32VarP(&rotation, "Rotation", "r", 0, "Rotation angle in degrees,eg: 0, 90, 180, 270")
	printCmd.Flags().Float32VarP(&scale, "Scale", "s", 1.0, "Scale factor of the image, don't use with fitToPaper")
	printCmd.Flags().BoolVarP(&fitToPaper, "fitToPaper", "a", false, "fit to paper")
	printCmd.MarkFlagRequired("printer")
	printCmd.MarkFlagRequired("file")
}
