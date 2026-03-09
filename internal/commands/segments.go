package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var segmentsCmd = &cobra.Command{
	Use:   "segments",
	Short: "Manage segments",
	Long: `List, get, and report on segments.

Examples:
  kv segments list
  kv segments get <ID>
  kv segments report <ID> --timeframe last_30_days`,
}

var segmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all segments",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "segments", nil)
			if err != nil {
				return err
			}
			return printData("segments.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "segments", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("segments.list", client.FlattenRaw(data, rawFlag))
	},
}

var segmentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get segment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "segments/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("segments.get", client.FlattenResponse(resp, rawFlag))
	},
}

var (
	segmentReportTimeframe    string
	segmentReportStart        string
	segmentReportEnd          string
	segmentReportStats        string
	segmentReportConversionID string
)

var segmentsReportCmd = &cobra.Command{
	Use:   "report <id>",
	Short: "Get segment performance report",
	Long: `Get segment analytics values.

Requires --conversion-metric-id (the metric ID for conversions, e.g. Placed Order).
Use 'kv metrics list' to find metric IDs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if segmentReportConversionID == "" {
			return fmt.Errorf("--conversion-metric-id is required (use 'kv metrics list' to find IDs)")
		}

		tf, err := parseTimeframe(segmentReportTimeframe, segmentReportStart, segmentReportEnd)
		if err != nil {
			return err
		}

		stats := []string{"opens", "open_rate", "clicks", "click_rate", "recipients", "delivered", "unsubscribes"}
		if segmentReportStats != "" {
			stats = strings.Split(segmentReportStats, ",")
		}

		body := jsonapiBody("segment-values-report", map[string]any{
			"statistics":           stats,
			"timeframe":            tf,
			"conversion_metric_id": segmentReportConversionID,
			"filter":               fmt.Sprintf("equals(segment_id,\"%s\")", args[0]),
		})

		resp, err := c.Post(ctx, "segment-values-reports", body)
		if err != nil {
			return err
		}
		return printData("segments.report", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	segmentsReportCmd.Flags().StringVar(&segmentReportTimeframe, "timeframe", "", "Predefined timeframe")
	segmentsReportCmd.Flags().StringVar(&segmentReportStart, "start", "", "Custom range start (ISO 8601)")
	segmentsReportCmd.Flags().StringVar(&segmentReportEnd, "end", "", "Custom range end (ISO 8601)")
	segmentsReportCmd.Flags().StringVar(&segmentReportStats, "stats", "", "Comma-separated statistics")
	segmentsReportCmd.Flags().StringVar(&segmentReportConversionID, "conversion-metric-id", "", "Conversion metric ID (required)")

	segmentsCmd.AddCommand(segmentsListCmd, segmentsGetCmd, segmentsReportCmd)
	rootCmd.AddCommand(segmentsCmd)
}
