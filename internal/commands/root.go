package commands

import (
	"github.com/nicolasacchi/kv/internal/client"
	"github.com/nicolasacchi/kv/internal/config"
	"github.com/nicolasacchi/kv/internal/output"
	"github.com/spf13/cobra"
)

var (
	version       = "dev"
	apiKeyFlag    string
	projectFlag   string
	revisionFlag  string
	jsonFlag      bool
	jqFlag        string
	rawFlag       bool
	verboseFlag   bool
	quietFlag     bool
	maxResultsFlag int
	noPaginateFlag bool
	outputDirFlag  string
)

var rootCmd = &cobra.Command{
	Use:   "kv",
	Short: "Klaviyo CLI — agentic-first command line tool for the Klaviyo API",
	Long: `kv is a CLI that covers the full Klaviyo API surface. Designed for
agentic tools like Claude Code and automation scripts.

Usage examples:
  kv campaigns list --json
  kv profiles get <ID>
  kv metrics aggregates <ID> --measurements count --interval day
  kv campaigns report <ID> --timeframe last_30_days`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func getClient(cmd *cobra.Command) (*client.Client, error) {
	apiKey, err := config.LoadAPIKey(apiKeyFlag, projectFlag)
	if err != nil {
		return nil, err
	}
	revision := config.LoadRevision(revisionFlag, projectFlag)
	return client.New(apiKey, "", revision, verboseFlag), nil
}

func isJSONMode() bool {
	return output.IsJSON(jsonFlag, jqFlag)
}

func printData(command string, data []byte) error {
	return output.PrintData(command, data, isJSONMode(), jqFlag, rawFlag)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Klaviyo API key (overrides KLAVIYO_API_KEY env var)")
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "Use a named project from config file")
	rootCmd.PersistentFlags().StringVar(&revisionFlag, "revision", "", "API revision header (default: "+client.DefaultRevision+")")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output (auto-enabled when stdout is not a TTY)")
	rootCmd.PersistentFlags().StringVar(&jqFlag, "jq", "", "Apply gjson path filter to JSON output")
	rootCmd.PersistentFlags().BoolVar(&rawFlag, "raw", false, "Output raw JSON:API response without flattening")
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Print request/response details to stderr")
	rootCmd.PersistentFlags().BoolVar(&quietFlag, "quiet", false, "Suppress non-error output")
	rootCmd.PersistentFlags().IntVar(&maxResultsFlag, "max-results", 0, "Maximum results to return (0 = unlimited)")
	rootCmd.PersistentFlags().BoolVar(&noPaginateFlag, "no-paginate", false, "Return first page only, don't auto-paginate")
	rootCmd.PersistentFlags().StringVar(&outputDirFlag, "output-dir", "", "Directory for file exports (default: current directory)")
}
