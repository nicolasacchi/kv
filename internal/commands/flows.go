package commands

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var flowStatusFilter string

var flowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "Manage flows",
	Long: `List, get, update, and report on flows.

Examples:
  kv flows list --status live
  kv flows get <ID>
  kv flows report <ID> --timeframe last_30_days`,
}

var flowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all flows",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		filter := filterEquals("status", flowStatusFilter)
		if filter != "" {
			params.Set("filter", filter)
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "flows", params)
			if err != nil {
				return err
			}
			return printData("flows.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "flows", params, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("flows.list", client.FlattenRaw(data, rawFlag))
	},
}

var flowsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get flow details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "flows/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("flows.get", client.FlattenResponse(resp, rawFlag))
	},
}

var flowsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update flow status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		if status == "" {
			return fmt.Errorf("--status is required (draft, live, manual)")
		}

		body := jsonapiBodyWithID("flow", args[0], map[string]any{
			"status": status,
		})

		resp, err := c.Patch(ctx, "flows/"+args[0], body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Flow updated.")
		return printData("flows.update", client.FlattenResponse(resp, rawFlag))
	},
}

var flowsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a flow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "flows/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Flow deleted.")
		return nil
	},
}

var (
	flowReportTimeframe    string
	flowReportStart        string
	flowReportEnd          string
	flowReportStats        string
	flowReportConversionID string
)

var flowsReportCmd = &cobra.Command{
	Use:   "report <id>",
	Short: "Get flow performance report",
	Long: `Get flow analytics values.

Requires --conversion-metric-id (the metric ID for conversions, e.g. Placed Order).
Use 'kv metrics list' to find metric IDs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if flowReportConversionID == "" {
			return fmt.Errorf("--conversion-metric-id is required (use 'kv metrics list' to find IDs)")
		}

		tf, err := parseTimeframe(flowReportTimeframe, flowReportStart, flowReportEnd)
		if err != nil {
			return err
		}

		stats := []string{"opens", "open_rate", "clicks", "click_rate", "recipients", "delivered", "unsubscribes"}
		if flowReportStats != "" {
			stats = strings.Split(flowReportStats, ",")
		}

		body := jsonapiBody("flow-values-report", map[string]any{
			"statistics":           stats,
			"timeframe":            tf,
			"conversion_metric_id": flowReportConversionID,
			"filter":               fmt.Sprintf("equals(flow_id,\"%s\")", args[0]),
		})

		resp, err := c.Post(ctx, "flow-values-reports", body)
		if err != nil {
			return err
		}
		return printData("flows.report", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	flowsListCmd.Flags().StringVar(&flowStatusFilter, "status", "", "Filter by status (draft, live, manual)")
	flowsUpdateCmd.Flags().String("status", "", "New status (draft, live, manual)")

	flowsReportCmd.Flags().StringVar(&flowReportTimeframe, "timeframe", "", "Predefined timeframe (e.g. last_30_days)")
	flowsReportCmd.Flags().StringVar(&flowReportStart, "start", "", "Custom range start (ISO 8601)")
	flowsReportCmd.Flags().StringVar(&flowReportEnd, "end", "", "Custom range end (ISO 8601)")
	flowsReportCmd.Flags().StringVar(&flowReportStats, "stats", "", "Comma-separated statistics")
	flowsReportCmd.Flags().StringVar(&flowReportConversionID, "conversion-metric-id", "", "Conversion metric ID (required)")

	flowsCmd.AddCommand(flowsListCmd, flowsGetCmd, flowsUpdateCmd, flowsDeleteCmd, flowsReportCmd)
	rootCmd.AddCommand(flowsCmd)
}
