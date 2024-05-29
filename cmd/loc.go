package cmd

import (
	"fmt"
	"mohh2-tools/internal/loc"

	"github.com/spf13/cobra"
)

// locCmd represents the loc command
var locCmd = &cobra.Command{
	Use:   "loc",
	Short: "Etract all strings from the game and convert strings to hashes",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("hash") {
			fmt.Printf("%08X\n", uint32(loc.GetHash(cmd.Flag("hash").Value.String())))
			return
		}

		paths := loc.FindDumps()
		for _, path := range paths {
			fmt.Println(path)
			loc.ExtractAllStrings(path)
		}
		fmt.Println("Strings extracted to the extracted directory")
	},
}

func init() {
	rootCmd.AddCommand(locCmd)
	locCmd.Flags().BoolP("extract", "e", false, "Extract all strings with their hashes and labels from the game to the extracted directory")
	locCmd.Flags().String("hash", "", "Convert a string to a hash")
}
