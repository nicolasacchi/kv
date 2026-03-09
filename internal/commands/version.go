package commands

import (
	"encoding/json"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := map[string]string{
			"version":    version,
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		}
		data, _ := json.Marshal(info)
		return printData("version", data)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
