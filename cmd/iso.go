package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	input string
)

// isoCmd represents the iso command
var isoCmd = &cobra.Command{
	Use:   "iso",
	Short: "Extract the contents of ISO/RVZ files or apply user defined patches",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Println("Input file not found")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(isoCmd)

	isoCmd.PersistentFlags().StringVarP(&input, "input", "i", "", "Input ISO|RVZ file")
	isoCmd.MarkPersistentFlagRequired("input")
}
