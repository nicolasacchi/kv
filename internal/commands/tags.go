package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags",
	Long: `List, get, create, assign, and remove tags.

Examples:
  kv tags list
  kv tags get <ID>
  kv tags create --name "Sale"
  kv tags assign <TAG_ID> --resource-type campaign --resource-id <ID>`,
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "tags", nil)
			if err != nil {
				return err
			}
			return printData("tags.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "tags", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("tags.list", client.FlattenRaw(data, rawFlag))
	},
}

var tagsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get tag details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "tags/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("tags.get", client.FlattenResponse(resp, rawFlag))
	},
}

var tagsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new tag",
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

		attrs := map[string]any{"name": name}

		groupID, _ := cmd.Flags().GetString("group-id")
		body := jsonapiBody("tag", attrs)
		if groupID != "" {
			bodyData := body["data"].(map[string]any)
			bodyData["relationships"] = map[string]any{
				"tag-group": jsonapiRelationship("tag-group", groupID),
			}
		}

		resp, err := c.Post(ctx, "tags", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Tag created.")
		return printData("tags.create", client.FlattenResponse(resp, rawFlag))
	},
}

var tagsAssignCmd = &cobra.Command{
	Use:   "assign <tag-id>",
	Short: "Assign a tag to a resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resourceType, _ := cmd.Flags().GetString("resource-type")
		resourceID, _ := cmd.Flags().GetString("resource-id")
		if resourceType == "" || resourceID == "" {
			return fmt.Errorf("--resource-type and --resource-id are required")
		}

		body := map[string]any{
			"data": map[string]any{
				"type": "tag",
				"id":   args[0],
				"relationships": map[string]any{
					resourceType + "s": map[string]any{
						"data": []map[string]any{
							{"type": resourceType, "id": resourceID},
						},
					},
				},
			},
		}

		path := fmt.Sprintf("tags/%s/relationships/%ss", args[0], resourceType)
		_, err = c.Post(ctx, path, body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Tag assigned.")
		return nil
	},
}

var tagsRemoveCmd = &cobra.Command{
	Use:   "remove <tag-id>",
	Short: "Remove a tag from a resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resourceType, _ := cmd.Flags().GetString("resource-type")
		resourceID, _ := cmd.Flags().GetString("resource-id")
		if resourceType == "" || resourceID == "" {
			return fmt.Errorf("--resource-type and --resource-id are required")
		}

		path := fmt.Sprintf("tags/%s/relationships/%ss", args[0], resourceType)
		err = c.Delete(ctx, path)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Tag removed.")
		return nil
	},
}

func init() {
	tagsCreateCmd.Flags().String("name", "", "Tag name (required)")
	tagsCreateCmd.Flags().String("group-id", "", "Tag group ID")

	tagsAssignCmd.Flags().String("resource-type", "", "Resource type (campaign, flow, list, segment)")
	tagsAssignCmd.Flags().String("resource-id", "", "Resource ID")

	tagsRemoveCmd.Flags().String("resource-type", "", "Resource type")
	tagsRemoveCmd.Flags().String("resource-id", "", "Resource ID")

	tagsCmd.AddCommand(tagsListCmd, tagsGetCmd, tagsCreateCmd, tagsAssignCmd, tagsRemoveCmd)
	rootCmd.AddCommand(tagsCmd)
}
