package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

var templatesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new template",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		attrs := map[string]any{
			"name": name,
		}

		editorType, _ := cmd.Flags().GetString("editor-type")
		if editorType != "" {
			attrs["editor_type"] = editorType
		}

		html, _ := cmd.Flags().GetString("html")
		if html != "" {
			if strings.HasPrefix(html, "@") {
				fileData, err := os.ReadFile(html[1:])
				if err != nil {
					return fmt.Errorf("read html file: %w", err)
				}
				html = string(fileData)
			}
			attrs["html"] = html
		}

		body := jsonapiBody("template", attrs)
		resp, err := c.Post(ctx, "templates", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Template created.")
		return printData("templates.create", client.FlattenResponse(resp, rawFlag))
	},
}

var templatesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		html, _ := cmd.Flags().GetString("html")

		attrs := map[string]any{}
		if name != "" {
			attrs["name"] = name
		}
		if html != "" {
			if strings.HasPrefix(html, "@") {
				fileData, err := os.ReadFile(html[1:])
				if err != nil {
					return fmt.Errorf("read html file: %w", err)
				}
				html = string(fileData)
			}
			attrs["html"] = html
		}

		if len(attrs) == 0 {
			return fmt.Errorf("at least one attribute to update is required (--name, --html)")
		}

		body := jsonapiBodyWithID("template", args[0], attrs)
		resp, err := c.Patch(ctx, "templates/"+args[0], body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Template updated.")
		return printData("templates.update", client.FlattenResponse(resp, rawFlag))
	},
}

var templatesCloneCmd = &cobra.Command{
	Use:   "clone <id>",
	Short: "Clone a template",
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

		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			attrs["name"] = name
		}

		body := jsonapiBody("template-clone", attrs)
		resp, err := c.Post(ctx, "template-clone", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Template cloned.")
		return printData("templates.clone", client.FlattenResponse(resp, rawFlag))
	},
}

var templatesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "templates/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Template deleted.")
		return nil
	},
}

func init() {
	templatesCreateCmd.Flags().String("name", "", "Template name (required)")
	templatesCreateCmd.Flags().String("html", "", "HTML content or @filepath to read from file")
	templatesCreateCmd.Flags().String("editor-type", "", "Editor type (CODE or DRAG_AND_DROP)")

	templatesUpdateCmd.Flags().String("name", "", "New template name")
	templatesUpdateCmd.Flags().String("html", "", "New HTML content or @filepath to read from file")

	templatesCloneCmd.Flags().String("name", "", "Name for the cloned template")

	templatesRenderCmd.Flags().String("context", "", "Render context as JSON string")

	templatesCmd.AddCommand(templatesListCmd, templatesGetCmd, templatesCreateCmd, templatesUpdateCmd, templatesCloneCmd, templatesDeleteCmd, templatesRenderCmd)
	rootCmd.AddCommand(templatesCmd)
}
