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

var printSetting = winspool.PrinterSetting{}

var imageSetting = winspool.ImageSetting{
	PPI:        96,    // windows default ppi
	Rotation:   0,     //rotation angle in degrees,eg: 0, 90, 180, 270
	Scale:      1.0,   // scale factor of the image. 1.0 means 100%.
	FitToPaper: false, // fit to paper
}

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
	jobID, err := winspool.ImageDataToPrinter(printerName, file_name, imageData, imageSetting, printSetting)
	if err != nil {
		return err
	}

	fmt.Println("jobID: ", jobID)

	return nil
}

func init() {
	rootCmd.AddCommand(printCmd)

	printCmd.Flags().StringVarP(&printType, "Type", "t", "IMAGE", "Data type send to printer：RAW，IMAGE，PDF")
	printCmd.Flags().StringVarP(&printerName, "Printer", "p", "", "Printer Name")
	printCmd.Flags().StringVarP(&filePath, "File", "f", "", "file path")

	printCmd.Flags().Int32VarP(&imageSetting.PPI, "Ppi", "i", 96, "Pixels per inch of the image")
	printCmd.Flags().Int32VarP(&imageSetting.Rotation, "Rotation", "r", 0, "Rotation angle in degrees,eg: 0, 90, 180, 270")
	printCmd.Flags().Float32VarP(&imageSetting.Scale, "Scale", "s", 1.0, "Scale factor of the image, don't use with fitToPaper")
	printCmd.Flags().BoolVarP(&imageSetting.FitToPaper, "FitToPaper", "a", false, "fit to paper")

	printCmd.Flags().StringVarP((*string)(&printSetting.PageOrientation), "Orientation", "o", "", "Page Orientation")
	printCmd.Flags().Int16VarP(&printSetting.Copy, "Copy", "c", 0, "Copies of the document to print")

	printCmd.MarkFlagRequired("printer")
	printCmd.MarkFlagRequired("file")
}
