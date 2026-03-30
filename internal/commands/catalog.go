package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Manage catalog items and variants",
	Long: `List, get, create, update, and delete catalog items and variants.

Examples:
  kv catalog items list
  kv catalog items get <ID>
  kv catalog items create --payload item.json
  kv catalog items update --payload item.json
  kv catalog items delete <ID>
  kv catalog variants list <ITEM_ID>
  kv catalog variants create --payload variant.json
  kv catalog variants delete <ID>`,
}

// Items subcommand group
var catalogItemsCmd = &cobra.Command{
	Use:   "items",
	Short: "Manage catalog items",
}

var catalogItemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all catalog items",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "catalog-items", nil)
			if err != nil {
				return err
			}
			return printData("catalog.items.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "catalog-items", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("catalog.items.list", client.FlattenRaw(data, rawFlag))
	},
}

var catalogItemsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get catalog item details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "catalog-items/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("catalog.items.get", client.FlattenResponse(resp, rawFlag))
	},
}

var catalogItemsCreateCmd = &cobra.Command{
	Use:   "create --payload <file>",
	Short: "Create a catalog item from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		payloadFile, _ := cmd.Flags().GetString("payload")
		if payloadFile == "" {
			return fmt.Errorf("--payload is required (path to JSON file)")
		}

		fileData, err := os.ReadFile(payloadFile)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var body map[string]any
		if err := json.Unmarshal(fileData, &body); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		resp, err := c.Post(ctx, "catalog-items", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Catalog item created.")
		return printData("catalog.items.create", client.FlattenResponse(resp, rawFlag))
	},
}

var catalogItemsUpdateCmd = &cobra.Command{
	Use:   "update --payload <file>",
	Short: "Update a catalog item from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		payloadFile, _ := cmd.Flags().GetString("payload")
		if payloadFile == "" {
			return fmt.Errorf("--payload is required (path to JSON file)")
		}

		fileData, err := os.ReadFile(payloadFile)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var body map[string]any
		if err := json.Unmarshal(fileData, &body); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		data, ok := body["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("file must contain a JSON:API data object with an id field")
		}
		id, ok := data["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("data.id is required for update")
		}

		resp, err := c.Patch(ctx, "catalog-items/"+id, body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Catalog item updated.")
		return printData("catalog.items.update", client.FlattenResponse(resp, rawFlag))
	},
}

var catalogItemsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a catalog item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "catalog-items/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Catalog item deleted.")
		return nil
	},
}

// Variants subcommand group
var catalogVariantsCmd = &cobra.Command{
	Use:   "variants",
	Short: "Manage catalog variants",
}

var catalogVariantsListCmd = &cobra.Command{
	Use:   "list <item-id>",
	Short: "List variants for a catalog item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("catalog-items/%s/variants", args[0])
		if noPaginateFlag {
			resp, err := c.Get(ctx, path, nil)
			if err != nil {
				return err
			}
			return printData("catalog.variants.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, path, nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("catalog.variants.list", client.FlattenRaw(data, rawFlag))
	},
}

var catalogVariantsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get variant details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "catalog-variants/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("catalog.variants.get", client.FlattenResponse(resp, rawFlag))
	},
}

var catalogVariantsCreateCmd = &cobra.Command{
	Use:   "create --payload <file>",
	Short: "Create a catalog variant from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		payloadFile, _ := cmd.Flags().GetString("payload")
		if payloadFile == "" {
			return fmt.Errorf("--payload is required (path to JSON file)")
		}

		fileData, err := os.ReadFile(payloadFile)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var body map[string]any
		if err := json.Unmarshal(fileData, &body); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		resp, err := c.Post(ctx, "catalog-variants", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Catalog variant created.")
		return printData("catalog.variants.create", client.FlattenResponse(resp, rawFlag))
	},
}

var catalogVariantsUpdateCmd = &cobra.Command{
	Use:   "update --payload <file>",
	Short: "Update a catalog variant from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		payloadFile, _ := cmd.Flags().GetString("payload")
		if payloadFile == "" {
			return fmt.Errorf("--payload is required (path to JSON file)")
		}

		fileData, err := os.ReadFile(payloadFile)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var body map[string]any
		if err := json.Unmarshal(fileData, &body); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		data, ok := body["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("file must contain a JSON:API data object with an id field")
		}
		id, ok := data["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("data.id is required for update")
		}

		resp, err := c.Patch(ctx, "catalog-variants/"+id, body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Catalog variant updated.")
		return printData("catalog.variants.update", client.FlattenResponse(resp, rawFlag))
	},
}

var catalogVariantsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a catalog variant",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "catalog-variants/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Catalog variant deleted.")
		return nil
	},
}

func init() {
	catalogItemsCreateCmd.Flags().String("payload", "", "Path to JSON file with item data")
	catalogItemsUpdateCmd.Flags().String("payload", "", "Path to JSON file with item data")

	catalogVariantsCreateCmd.Flags().String("payload", "", "Path to JSON file with variant data")
	catalogVariantsUpdateCmd.Flags().String("payload", "", "Path to JSON file with variant data")

	catalogItemsCmd.AddCommand(catalogItemsListCmd, catalogItemsGetCmd, catalogItemsCreateCmd, catalogItemsUpdateCmd, catalogItemsDeleteCmd)
	catalogVariantsCmd.AddCommand(catalogVariantsListCmd, catalogVariantsGetCmd, catalogVariantsCreateCmd, catalogVariantsUpdateCmd, catalogVariantsDeleteCmd)
	catalogCmd.AddCommand(catalogItemsCmd, catalogVariantsCmd)
	rootCmd.AddCommand(catalogCmd)
}
