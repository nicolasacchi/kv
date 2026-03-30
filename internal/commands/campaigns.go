package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var (
	campaignStatusFilter  string
	campaignChannelFilter string
)

var campaignsCmd = &cobra.Command{
	Use:   "campaigns",
	Short: "Manage campaigns",
	Long: `List, get, create, and report on campaigns.

Examples:
  kv campaigns list --status sent
  kv campaigns get <ID>
  kv campaigns report <ID> --timeframe last_30_days`,
}

var campaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		channel := campaignChannelFilter
		if channel == "" {
			channel = "email" // Klaviyo requires a channel filter
		}
		filter := buildFilter(
			filterEquals("messages.channel", channel),
		)
		if campaignStatusFilter != "" {
			filter = buildFilter(filter, filterEquals("status", campaignStatusFilter))
		}
		if filter != "" {
			params.Set("filter", filter)
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "campaigns", params)
			if err != nil {
				return err
			}
			return printData("campaigns.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "campaigns", params, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("campaigns.list", client.FlattenRaw(data, rawFlag))
	},
}

var campaignsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get campaign details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "campaigns/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("campaigns.get", client.FlattenResponse(resp, rawFlag))
	},
}

var campaignsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		channel, _ := cmd.Flags().GetString("channel")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		attrs := map[string]any{
			"name": name,
		}
		if channel != "" {
			attrs["channel"] = channel
		}

		body := jsonapiBody("campaign", attrs)
		resp, err := c.Post(ctx, "campaigns", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Campaign created.")
		return printData("campaigns.create", client.FlattenResponse(resp, rawFlag))
	},
}

var campaignsPutCmd = &cobra.Command{
	Use:   "put <file>",
	Short: "Update a campaign from a JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		fileData, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var body map[string]any
		if err := json.Unmarshal(fileData, &body); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		// Extract ID from the data object
		data, ok := body["data"].(map[string]any)
		if !ok {
			return fmt.Errorf("file must contain a JSON:API data object with an id field")
		}
		id, ok := data["id"].(string)
		if !ok || id == "" {
			return fmt.Errorf("data.id is required for update")
		}

		resp, err := c.Patch(ctx, "campaigns/"+id, body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Campaign updated.")
		return printData("campaigns.put", client.FlattenResponse(resp, rawFlag))
	},
}

var campaignsCloneCmd = &cobra.Command{
	Use:   "clone <id>",
	Short: "Clone a campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		attrs := map[string]any{
			"campaign_id": args[0],
		}

		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			attrs["name"] = name
		}

		body := jsonapiBody("campaign-clone", attrs)
		resp, err := c.Post(ctx, "campaign-clone", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Campaign cloned.")
		return printData("campaigns.clone", client.FlattenResponse(resp, rawFlag))
	},
}

var campaignsSendCmd = &cobra.Command{
	Use:   "send <id>",
	Short: "Send a campaign",
	Long:  `Send a campaign immediately or schedule it for later. WARNING: This sends real emails to recipients.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		sendAt, _ := cmd.Flags().GetString("send-at")
		if sendAt != "" {
			fmt.Fprintf(os.Stderr, "WARNING: Scheduling campaign %s for %s.\n", args[0], sendAt)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Sending campaign %s immediately.\n", args[0])
		}

		body := map[string]any{
			"data": map[string]any{
				"type": "campaign-send-job",
			},
		}

		if sendAt != "" {
			bodyData := body["data"].(map[string]any)
			bodyData["attributes"] = map[string]any{
				"send_at": sendAt,
			}
		}

		// Add campaign relationship
		bodyData := body["data"].(map[string]any)
		bodyData["relationships"] = map[string]any{
			"campaign": jsonapiRelationship("campaign", args[0]),
		}

		resp, err := c.Post(ctx, "campaign-send-jobs", body)
		if err != nil {
			return err
		}
		if sendAt != "" {
			fmt.Fprintln(os.Stderr, "Campaign scheduled.")
		} else {
			fmt.Fprintln(os.Stderr, "Campaign send job created.")
		}
		return printData("campaigns.send", client.FlattenResponse(resp, rawFlag))
	},
}

var campaignsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "campaigns/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Campaign deleted.")
		return nil
	},
}

var (
	campaignReportTimeframe    string
	campaignReportStart        string
	campaignReportEnd          string
	campaignReportStats        string
	campaignReportConversionID string
)

var campaignsReportCmd = &cobra.Command{
	Use:   "report <id>",
	Short: "Get campaign performance report",
	Long: `Get campaign analytics values.

Requires --conversion-metric-id (the metric ID for conversions, e.g. Placed Order).
Use 'kv metrics list' to find metric IDs.

Available statistics:
  opens, opens_unique, open_rate, clicks, clicks_unique, click_rate,
  click_to_open_rate, recipients, delivered, delivery_rate, bounced,
  bounce_rate, unsubscribes, unsubscribe_rate, spam_complaints,
  conversion_rate, conversions, conversion_value, revenue_per_recipient`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if campaignReportConversionID == "" {
			return fmt.Errorf("--conversion-metric-id is required (use 'kv metrics list' to find IDs)")
		}

		tf, err := parseTimeframe(campaignReportTimeframe, campaignReportStart, campaignReportEnd)
		if err != nil {
			return err
		}

		stats := []string{"opens", "open_rate", "clicks", "click_rate", "recipients", "delivered", "unsubscribes"}
		if campaignReportStats != "" {
			stats = strings.Split(campaignReportStats, ",")
		}

		body := jsonapiBody("campaign-values-report", map[string]any{
			"statistics":           stats,
			"timeframe":            tf,
			"conversion_metric_id": campaignReportConversionID,
			"filter":               fmt.Sprintf("equals(campaign_id,\"%s\")", args[0]),
		})

		resp, err := c.Post(ctx, "campaign-values-reports", body)
		if err != nil {
			return err
		}
		return printData("campaigns.report", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	campaignsListCmd.Flags().StringVar(&campaignStatusFilter, "status", "", "Filter by status (draft, scheduled, sent)")
	campaignsListCmd.Flags().StringVar(&campaignChannelFilter, "channel", "", "Filter by channel (email, sms)")

	campaignsCreateCmd.Flags().String("name", "", "Campaign name (required)")
	campaignsCreateCmd.Flags().String("channel", "email", "Campaign channel (email, sms)")

	campaignsReportCmd.Flags().StringVar(&campaignReportTimeframe, "timeframe", "", "Predefined timeframe (e.g. last_30_days)")
	campaignsReportCmd.Flags().StringVar(&campaignReportStart, "start", "", "Custom range start (ISO 8601)")
	campaignsReportCmd.Flags().StringVar(&campaignReportEnd, "end", "", "Custom range end (ISO 8601)")
	campaignsReportCmd.Flags().StringVar(&campaignReportStats, "stats", "", "Comma-separated statistics (default: common set)")
	campaignsReportCmd.Flags().StringVar(&campaignReportConversionID, "conversion-metric-id", "", "Conversion metric ID (required, e.g. Placed Order metric)")

	campaignsCloneCmd.Flags().String("name", "", "Name for the cloned campaign")
	campaignsSendCmd.Flags().String("send-at", "", "Schedule send time (ISO 8601, omit for immediate)")

	campaignsCmd.AddCommand(campaignsListCmd, campaignsGetCmd, campaignsCreateCmd, campaignsPutCmd, campaignsCloneCmd, campaignsSendCmd, campaignsDeleteCmd, campaignsReportCmd)
	rootCmd.AddCommand(campaignsCmd)
}
