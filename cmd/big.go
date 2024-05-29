package cmd

import (
	"fmt"
	"mohh2-tools/internal/big"
	"mohh2-tools/internal/utils"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// bigCmd represents the big command
var bigCmd = &cobra.Command{
	Use:   "big",
	Short: "Extracts .big and .viv files in the game_files directory and its subdirectories",
	Long:  `Extracts .big and .viv files in the game_files directory and its subdirectories using quickbms 0.12.0 and ea_big4.bms 0.1.3 to extract the files`,
	Run: func(cmd *cobra.Command, args []string) {
		var tools = []string{"quickbms.exe", "ea_big4.bms"}
		for _, tool := range tools {
			if _, err := os.Stat(filepath.Join("tools", tool)); os.IsNotExist(err) {
				utils.DownloadQuickBMS()
				break
			}
		}
		if _, err := os.Stat(utils.GameDir()); os.IsNotExist(err) {
			fmt.Printf("%s directory not found\n", utils.GameDir())
			return
		}
		fmt.Println("Extracting...")
		big.ExtractBigVivFiles()
		fmt.Println("Done!")
	},
}

func init() {
	rootCmd.AddCommand(bigCmd)
}
