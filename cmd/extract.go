package cmd

import (
	"mohh2-tools/internal/extract"

	"github.com/spf13/cobra"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extracts the contents of the ISO to the game_files directory",
	Run: func(cmd *cobra.Command, args []string) {
		extract.ExtractToGameDir(input)
	},
}

func init() {
	isoCmd.AddCommand(extractCmd)
}
