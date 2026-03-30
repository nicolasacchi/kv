package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage events",
	Long: `List, get, and create events.

Examples:
  kv events list --metric-id <ID>
  kv events get <ID>
  kv events create --metric-name "Placed Order" --profile-email user@example.com`,
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		metricID, _ := cmd.Flags().GetString("metric-id")
		profileID, _ := cmd.Flags().GetString("profile-id")
		since, _ := cmd.Flags().GetString("since")
		until, _ := cmd.Flags().GetString("until")

		filter := buildFilter(
			filterEquals("metric_id", metricID),
			filterEquals("profile_id", profileID),
			filterGreaterOrEqual("datetime", since),
			filterLessOrEqual("datetime", until),
		)
		if filter != "" {
			params.Set("filter", filter)
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "events", params)
			if err != nil {
				return err
			}
			return printData("events.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "events", params, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("events.list", client.FlattenRaw(data, rawFlag))
	},
}

var eventsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get event details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "events/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("events.get", client.FlattenResponse(resp, rawFlag))
	},
}

var eventsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom event",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		metricName, _ := cmd.Flags().GetString("metric-name")
		profileEmail, _ := cmd.Flags().GetString("profile-email")
		propertiesJSON, _ := cmd.Flags().GetString("properties")

		if metricName == "" {
			return fmt.Errorf("--metric-name is required")
		}
		if profileEmail == "" {
			return fmt.Errorf("--profile-email is required")
		}

		attrs := map[string]any{
			"metric": map[string]any{
				"data": map[string]any{
					"type":       "metric",
					"attributes": map[string]any{"name": metricName},
				},
			},
			"profile": map[string]any{
				"data": map[string]any{
					"type":       "profile",
					"attributes": map[string]any{"email": profileEmail},
				},
			},
		}

		if propertiesJSON != "" {
			var props map[string]any
			if err := json.Unmarshal([]byte(propertiesJSON), &props); err != nil {
				return fmt.Errorf("invalid --properties JSON: %w", err)
			}
			attrs["properties"] = props
		}

		body := jsonapiBody("event", attrs)
		resp, err := c.Post(ctx, "events", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Event created.")
		return printData("events.create", client.FlattenResponse(resp, rawFlag))
	},
}

var eventsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create --payload <file>",
	Short: "Bulk create events from a JSON file",
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

		resp, err := c.Post(ctx, "event-bulk-create-jobs", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Bulk event creation job submitted.")
		return printData("events.bulk-create", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	eventsListCmd.Flags().String("metric-id", "", "Filter by metric ID")
	eventsListCmd.Flags().String("profile-id", "", "Filter by profile ID")
	eventsListCmd.Flags().String("since", "", "Filter events after this datetime (ISO 8601)")
	eventsListCmd.Flags().String("until", "", "Filter events before this datetime (ISO 8601)")

	eventsCreateCmd.Flags().String("metric-name", "", "Metric name (required)")
	eventsCreateCmd.Flags().String("profile-email", "", "Profile email (required)")
	eventsCreateCmd.Flags().String("properties", "", "Event properties as JSON string")

	eventsBulkCreateCmd.Flags().String("payload", "", "Path to JSON file with bulk event data")

	eventsCmd.AddCommand(eventsListCmd, eventsGetCmd, eventsCreateCmd, eventsBulkCreateCmd)
	rootCmd.AddCommand(eventsCmd)
}
