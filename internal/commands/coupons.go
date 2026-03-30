package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var couponsCmd = &cobra.Command{
	Use:   "coupons",
	Short: "Manage coupons and coupon codes",
	Long: `List, get, create, update, and delete coupons and coupon codes.

Examples:
  kv coupons list
  kv coupons get <ID>
  kv coupons create --external-id SUMMER2024
  kv coupons codes list <COUPON_ID>
  kv coupons codes create <COUPON_ID> --unique-code SAVE10`,
}

// Coupons subcommands

var couponsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all coupons",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "coupons", nil)
			if err != nil {
				return err
			}
			return printData("coupons.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "coupons", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("coupons.list", client.FlattenRaw(data, rawFlag))
	},
}

var couponsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get coupon details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "coupons/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("coupons.get", client.FlattenResponse(resp, rawFlag))
	},
}

var couponsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new coupon",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		externalID, _ := cmd.Flags().GetString("external-id")
		if externalID == "" {
			return fmt.Errorf("--external-id is required")
		}

		attrs := map[string]any{
			"external_id": externalID,
		}

		description, _ := cmd.Flags().GetString("description")
		if description != "" {
			attrs["description"] = description
		}

		body := jsonapiBody("coupon", attrs)
		resp, err := c.Post(ctx, "coupons", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Coupon created.")
		return printData("coupons.create", client.FlattenResponse(resp, rawFlag))
	},
}

var couponsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		attrs := map[string]any{}

		description, _ := cmd.Flags().GetString("description")
		if description != "" {
			attrs["description"] = description
		}

		if len(attrs) == 0 {
			return fmt.Errorf("at least one attribute to update is required (--description)")
		}

		body := jsonapiBodyWithID("coupon", args[0], attrs)
		resp, err := c.Patch(ctx, "coupons/"+args[0], body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Coupon updated.")
		return printData("coupons.update", client.FlattenResponse(resp, rawFlag))
	},
}

var couponsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "coupons/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Coupon deleted.")
		return nil
	},
}

// Coupon codes subcommand group

var couponsCodesCmd = &cobra.Command{
	Use:   "codes",
	Short: "Manage coupon codes",
}

var couponsCodesListCmd = &cobra.Command{
	Use:   "list <coupon-id>",
	Short: "List coupon codes for a coupon",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("coupons/%s/coupon-codes", args[0])
		if noPaginateFlag {
			resp, err := c.Get(ctx, path, nil)
			if err != nil {
				return err
			}
			return printData("coupons.codes.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, path, nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("coupons.codes.list", client.FlattenRaw(data, rawFlag))
	},
}

var couponsCodesGetCmd = &cobra.Command{
	Use:   "get <code-id>",
	Short: "Get coupon code details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "coupon-codes/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("coupons.codes.get", client.FlattenResponse(resp, rawFlag))
	},
}

var couponsCodesCreateCmd = &cobra.Command{
	Use:   "create <coupon-id>",
	Short: "Create a coupon code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		uniqueCode, _ := cmd.Flags().GetString("unique-code")
		if uniqueCode == "" {
			return fmt.Errorf("--unique-code is required")
		}

		attrs := map[string]any{
			"unique_code": uniqueCode,
		}

		expiresAt, _ := cmd.Flags().GetString("expires-at")
		if expiresAt != "" {
			attrs["expires_at"] = expiresAt
		}

		body := jsonapiBody("coupon-code", attrs)
		bodyData := body["data"].(map[string]any)
		bodyData["relationships"] = map[string]any{
			"coupon": jsonapiRelationship("coupon", args[0]),
		}

		resp, err := c.Post(ctx, "coupon-codes", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Coupon code created.")
		return printData("coupons.codes.create", client.FlattenResponse(resp, rawFlag))
	},
}

var couponsCodesUpdateCmd = &cobra.Command{
	Use:   "update <code-id>",
	Short: "Update a coupon code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		attrs := map[string]any{}

		status, _ := cmd.Flags().GetString("status")
		if status != "" {
			attrs["status"] = status
		}
		expiresAt, _ := cmd.Flags().GetString("expires-at")
		if expiresAt != "" {
			attrs["expires_at"] = expiresAt
		}

		if len(attrs) == 0 {
			return fmt.Errorf("at least one attribute to update is required (--status, --expires-at)")
		}

		body := jsonapiBodyWithID("coupon-code", args[0], attrs)
		resp, err := c.Patch(ctx, "coupon-codes/"+args[0], body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Coupon code updated.")
		return printData("coupons.codes.update", client.FlattenResponse(resp, rawFlag))
	},
}

var couponsCodesDeleteCmd = &cobra.Command{
	Use:   "delete <code-id>",
	Short: "Delete a coupon code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "coupon-codes/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Coupon code deleted.")
		return nil
	},
}

func init() {
	couponsCreateCmd.Flags().String("external-id", "", "External ID for the coupon (required)")
	couponsCreateCmd.Flags().String("description", "", "Coupon description")

	couponsUpdateCmd.Flags().String("description", "", "New coupon description")

	couponsCodesCreateCmd.Flags().String("unique-code", "", "Unique coupon code (required)")
	couponsCodesCreateCmd.Flags().String("expires-at", "", "Expiration datetime (ISO 8601)")

	couponsCodesUpdateCmd.Flags().String("status", "", "Coupon code status")
	couponsCodesUpdateCmd.Flags().String("expires-at", "", "New expiration datetime (ISO 8601)")

	couponsCodesCmd.AddCommand(couponsCodesListCmd, couponsCodesGetCmd, couponsCodesCreateCmd, couponsCodesUpdateCmd, couponsCodesDeleteCmd)
	couponsCmd.AddCommand(couponsListCmd, couponsGetCmd, couponsCreateCmd, couponsUpdateCmd, couponsDeleteCmd, couponsCodesCmd)
	rootCmd.AddCommand(couponsCmd)
}
