package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhooks",
	Long: `List, get, create, and delete webhooks.

Examples:
  kv webhooks list
  kv webhooks get <ID>
  kv webhooks create --url https://example.com/hook --events placed_order,ordered_product
  kv webhooks delete <ID>`,
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "webhooks", nil)
			if err != nil {
				return err
			}
			return printData("webhooks.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "webhooks", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("webhooks.list", client.FlattenRaw(data, rawFlag))
	},
}

var webhooksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get webhook details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "webhooks/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("webhooks.get", client.FlattenResponse(resp, rawFlag))
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		urlFlag, _ := cmd.Flags().GetString("url")
		eventsFlag, _ := cmd.Flags().GetString("events")
		secretFlag, _ := cmd.Flags().GetString("secret")

		if urlFlag == "" {
			return fmt.Errorf("--url is required")
		}
		if eventsFlag == "" {
			return fmt.Errorf("--events is required")
		}

		attrs := map[string]any{
			"endpoint_url": urlFlag,
			"events":       strings.Split(eventsFlag, ","),
		}
		if secretFlag != "" {
			attrs["secret"] = secretFlag
		}

		body := jsonapiBody("webhook", attrs)
		resp, err := c.Post(ctx, "webhooks", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Webhook created.")
		return printData("webhooks.create", client.FlattenResponse(resp, rawFlag))
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "webhooks/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Webhook deleted.")
		return nil
	},
}

func init() {
	webhooksCreateCmd.Flags().String("url", "", "Webhook endpoint URL (required)")
	webhooksCreateCmd.Flags().String("events", "", "Comma-separated event types (required)")
	webhooksCreateCmd.Flags().String("secret", "", "Webhook signing secret")

	webhooksCmd.AddCommand(webhooksListCmd, webhooksGetCmd, webhooksCreateCmd, webhooksDeleteCmd)
	rootCmd.AddCommand(webhooksCmd)
}
