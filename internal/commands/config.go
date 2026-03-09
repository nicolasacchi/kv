package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/config"
	"github.com/spf13/cobra"
)

var (
	configAddKey      string
	configAddRevision string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage project configurations",
	Long: `Add, remove, list, and switch between Klaviyo project configurations.

Projects are stored in ~/.config/kv/config.toml.

Examples:
  kv config add production --api-key pk_abc123
  kv config list
  kv config use production
  kv config current`,
}

var configAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update a project configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if configAddKey == "" {
			return fmt.Errorf("--api-key is required")
		}
		if err := config.AddProject(name, configAddKey, configAddRevision); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Project %q added.\n", name)
		result := map[string]string{"project": name, "status": "added"}
		data, _ := json.Marshal(result)
		return printData("config.add", data)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a project configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.RemoveProject(name); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Project %q removed.\n", name)
		result := map[string]string{"project": name, "status": "removed"}
		data, _ := json.Marshal(result)
		return printData("config.remove", data)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			data, _ := json.Marshal([]any{})
			return printData("config.list", data)
		}

		var rows []map[string]any
		if cfg.Projects != nil {
			for name, p := range cfg.Projects {
				rows = append(rows, map[string]any{
					"name":     name,
					"api_key":  config.MaskKey(p.APIKey),
					"revision": p.Revision,
					"default":  name == cfg.DefaultProject,
				})
			}
		}

		if rows == nil {
			rows = []map[string]any{}
		}
		data, _ := json.Marshal(rows)
		return printData("config.list", data)
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.SetDefaultProject(name); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Default project set to %q.\n", name)
		result := map[string]string{"project": name, "status": "default"}
		data, _ := json.Marshal(result)
		return printData("config.use", data)
	},
}

var configCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active project and resolved API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			return fmt.Errorf("no config file found")
		}

		var name string
		var source string
		if projectFlag != "" {
			name = projectFlag
			source = "flag"
		} else if cfg.DefaultProject != "" {
			name = cfg.DefaultProject
			source = "default"
		} else {
			return fmt.Errorf("no project configured")
		}

		result := map[string]string{"project": name, "source": source}
		data, _ := json.Marshal(result)
		return printData("config.current", data)
	},
}

func init() {
	configAddCmd.Flags().StringVar(&configAddKey, "api-key", "", "Klaviyo API key (required)")
	configAddCmd.Flags().StringVar(&configAddRevision, "revision", "", "API revision for this project")

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configCurrentCmd)
	rootCmd.AddCommand(configCmd)
}
