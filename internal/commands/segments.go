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

var segmentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new segment",
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

		definitionJSON, _ := cmd.Flags().GetString("definition")
		if definitionJSON != "" {
			var def map[string]any
			if err := json.Unmarshal([]byte(definitionJSON), &def); err != nil {
				return fmt.Errorf("invalid --definition JSON: %w", err)
			}
			attrs["definition"] = def
		}

		body := jsonapiBody("segment", attrs)
		resp, err := c.Post(ctx, "segments", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Segment created.")
		return printData("segments.create", client.FlattenResponse(resp, rawFlag))
	},
}

var segmentsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a segment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		definitionJSON, _ := cmd.Flags().GetString("definition")

		attrs := map[string]any{}
		if name != "" {
			attrs["name"] = name
		}
		if definitionJSON != "" {
			var def map[string]any
			if err := json.Unmarshal([]byte(definitionJSON), &def); err != nil {
				return fmt.Errorf("invalid --definition JSON: %w", err)
			}
			attrs["definition"] = def
		}

		if len(attrs) == 0 {
			return fmt.Errorf("at least one attribute to update is required (--name, --definition)")
		}

		body := jsonapiBodyWithID("segment", args[0], attrs)
		resp, err := c.Patch(ctx, "segments/"+args[0], body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Segment updated.")
		return printData("segments.update", client.FlattenResponse(resp, rawFlag))
	},
}

var segmentsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a segment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		err = c.Delete(ctx, "segments/"+args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Segment deleted.")
		return nil
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

	segmentsCreateCmd.Flags().String("name", "", "Segment name (required)")
	segmentsCreateCmd.Flags().String("definition", "", "Segment definition as JSON string")

	segmentsUpdateCmd.Flags().String("name", "", "New segment name")
	segmentsUpdateCmd.Flags().String("definition", "", "New segment definition as JSON string")

	segmentsCmd.AddCommand(segmentsListCmd, segmentsGetCmd, segmentsCreateCmd, segmentsUpdateCmd, segmentsDeleteCmd, segmentsReportCmd)
	rootCmd.AddCommand(segmentsCmd)
}
