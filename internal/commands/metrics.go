package commands

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Manage metrics and query aggregates",
	Long: `List metrics, get details, and query metric aggregates.

Examples:
  kv metrics list
  kv metrics get <ID>
  kv metrics aggregates <ID> --measurements count --interval day --start 2024-01-01 --end 2024-01-31`,
}

var metricsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		integration, _ := cmd.Flags().GetString("integration")
		if integration != "" {
			params.Set("filter", filterEquals("integration.name", integration))
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "metrics", params)
			if err != nil {
				return err
			}
			return printData("metrics.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "metrics", params, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("metrics.list", client.FlattenRaw(data, rawFlag))
	},
}

var metricsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get metric details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "metrics/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("metrics.get", client.FlattenResponse(resp, rawFlag))
	},
}

var (
	aggMeasurements string
	aggInterval     string
	aggStart        string
	aggEnd          string
	aggBy           string
	aggFilter       string
	aggTimezone     string
)

var metricsAggregatesCmd = &cobra.Command{
	Use:   "aggregates <metric-id>",
	Short: "Query metric aggregates",
	Long: `Query and aggregate event data with filtering and grouping.

Measurements: count, sum_value, unique, average_value, min_value, max_value
Intervals: day, week, month
Group by: $flow, $campaign, $message, $variation, $attributed_channel

Examples:
  kv metrics aggregates <ID> --measurements count,unique --interval day --start 2024-06-01 --end 2024-06-30
  kv metrics aggregates <ID> --measurements count --by '$flow' --interval week`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if aggStart == "" || aggEnd == "" {
			return fmt.Errorf("--start and --end are required for metric aggregates")
		}

		measurements := []string{"count"}
		if aggMeasurements != "" {
			measurements = strings.Split(aggMeasurements, ",")
		}

		interval := "day"
		if aggInterval != "" {
			interval = aggInterval
		}

		timezone := "UTC"
		if aggTimezone != "" {
			timezone = aggTimezone
		}

		attrs := map[string]any{
			"metric_id":    args[0],
			"measurements": measurements,
			"interval":     interval,
			"page_size":    500,
			"timezone":     timezone,
			"filter": []string{
				fmt.Sprintf("greater-or-equal(datetime,%s)", aggStart),
				fmt.Sprintf("less-than(datetime,%s)", aggEnd),
			},
		}

		if aggBy != "" {
			attrs["by"] = strings.Split(aggBy, ",")
		}

		if aggFilter != "" {
			// Append custom filter to existing date filters
			filters := attrs["filter"].([]string)
			filters = append(filters, aggFilter)
			attrs["filter"] = filters
		}

		body := jsonapiBody("metric-aggregate", attrs)
		resp, err := c.Post(ctx, "metric-aggregates", body)
		if err != nil {
			return err
		}

		if rawFlag {
			return printData("metrics.aggregates", resp.Data)
		}
		return printData("metrics.aggregates", client.FlattenAggregates(resp))
	},
}

func init() {
	metricsListCmd.Flags().String("integration", "", "Filter by integration name")

	metricsAggregatesCmd.Flags().StringVar(&aggMeasurements, "measurements", "count", "Comma-separated measurements (count, sum_value, unique, average_value, min_value, max_value)")
	metricsAggregatesCmd.Flags().StringVar(&aggInterval, "interval", "day", "Aggregation interval (day, week, month)")
	metricsAggregatesCmd.Flags().StringVar(&aggStart, "start", "", "Date range start (ISO 8601, required)")
	metricsAggregatesCmd.Flags().StringVar(&aggEnd, "end", "", "Date range end (ISO 8601, required)")
	metricsAggregatesCmd.Flags().StringVar(&aggBy, "by", "", "Comma-separated grouping dimensions ($flow, $campaign, etc.)")
	metricsAggregatesCmd.Flags().StringVar(&aggFilter, "filter", "", "Additional filter expression")
	metricsAggregatesCmd.Flags().StringVar(&aggTimezone, "timezone", "UTC", "Timezone for aggregation")

	metricsCmd.AddCommand(metricsListCmd, metricsGetCmd, metricsAggregatesCmd)
	rootCmd.AddCommand(metricsCmd)
}
