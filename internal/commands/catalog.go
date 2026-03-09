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
	Long: `List, get, and create catalog items and variants.

Examples:
  kv catalog items list
  kv catalog items get <ID>
  kv catalog variants list <ITEM_ID>`,
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

func init() {
	catalogItemsCreateCmd.Flags().String("payload", "", "Path to JSON file with item data")

	catalogItemsCmd.AddCommand(catalogItemsListCmd, catalogItemsGetCmd, catalogItemsCreateCmd)
	catalogVariantsCmd.AddCommand(catalogVariantsListCmd, catalogVariantsGetCmd)
	catalogCmd.AddCommand(catalogItemsCmd, catalogVariantsCmd)
	rootCmd.AddCommand(catalogCmd)
}
