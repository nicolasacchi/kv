package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage templates",
	Long: `List, get, and render templates.

Examples:
  kv templates list
  kv templates get <ID>
  kv templates render <ID>`,
}

var templatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "templates", nil)
			if err != nil {
				return err
			}
			return printData("templates.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "templates", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("templates.list", client.FlattenRaw(data, rawFlag))
	},
}

var templatesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get template details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "templates/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("templates.get", client.FlattenResponse(resp, rawFlag))
	},
}

var templatesRenderCmd = &cobra.Command{
	Use:   "render <id>",
	Short: "Render a template with optional context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		attrs := map[string]any{
			"id": args[0],
		}

		contextJSON, _ := cmd.Flags().GetString("context")
		if contextJSON != "" {
			var ctxData map[string]any
			if err := json.Unmarshal([]byte(contextJSON), &ctxData); err != nil {
				return fmt.Errorf("invalid --context JSON: %w", err)
			}
			attrs["context"] = ctxData
		}

		body := jsonapiBody("template", attrs)
		resp, err := c.Post(ctx, "template-render", body)
		if err != nil {
			return err
		}
		return printData("templates.render", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	templatesRenderCmd.Flags().String("context", "", "Render context as JSON string")

	templatesCmd.AddCommand(templatesListCmd, templatesGetCmd, templatesRenderCmd)
	rootCmd.AddCommand(templatesCmd)
}
