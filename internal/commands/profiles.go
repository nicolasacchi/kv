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

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage profiles",
	Long: `List, get, create, update, and suppress profiles.

Examples:
  kv profiles list --email user@example.com
  kv profiles get <ID>
  kv profiles create --email user@example.com --first-name Jane
  kv profiles suppress <ID>`,
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		email, _ := cmd.Flags().GetString("email")
		phone, _ := cmd.Flags().GetString("phone")

		filter := buildFilter(
			filterEquals("email", email),
			filterEquals("phone_number", phone),
		)
		if filter != "" {
			params.Set("filter", filter)
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "profiles", params)
			if err != nil {
				return err
			}
			return printData("profiles.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "profiles", params, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("profiles.list", client.FlattenRaw(data, rawFlag))
	},
}

var profilesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get profile details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "profiles/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("profiles.get", client.FlattenResponse(resp, rawFlag))
	},
}

var profilesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("email")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		phone, _ := cmd.Flags().GetString("phone")
		propertiesJSON, _ := cmd.Flags().GetString("properties")

		if email == "" {
			return fmt.Errorf("--email is required")
		}

		attrs := map[string]any{
			"email": email,
		}
		if firstName != "" {
			attrs["first_name"] = firstName
		}
		if lastName != "" {
			attrs["last_name"] = lastName
		}
		if phone != "" {
			attrs["phone_number"] = phone
		}
		if propertiesJSON != "" {
			var props map[string]any
			if err := json.Unmarshal([]byte(propertiesJSON), &props); err != nil {
				return fmt.Errorf("invalid --properties JSON: %w", err)
			}
			attrs["properties"] = props
		}

		body := jsonapiBody("profile", attrs)
		resp, err := c.Post(ctx, "profiles", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Profile created.")
		return printData("profiles.create", client.FlattenResponse(resp, rawFlag))
	},
}

var profilesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update profile attributes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		phone, _ := cmd.Flags().GetString("phone")
		propertiesJSON, _ := cmd.Flags().GetString("properties")

		attrs := map[string]any{}
		if firstName != "" {
			attrs["first_name"] = firstName
		}
		if lastName != "" {
			attrs["last_name"] = lastName
		}
		if phone != "" {
			attrs["phone_number"] = phone
		}
		if propertiesJSON != "" {
			var props map[string]any
			if err := json.Unmarshal([]byte(propertiesJSON), &props); err != nil {
				return fmt.Errorf("invalid --properties JSON: %w", err)
			}
			attrs["properties"] = props
		}

		if len(attrs) == 0 {
			return fmt.Errorf("at least one attribute to update is required")
		}

		body := jsonapiBodyWithID("profile", args[0], attrs)
		resp, err := c.Patch(ctx, "profiles/"+args[0], body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Profile updated.")
		return printData("profiles.update", client.FlattenResponse(resp, rawFlag))
	},
}

var profilesSuppressCmd = &cobra.Command{
	Use:   "suppress <id>",
	Short: "Suppress a profile (GDPR/unsubscribe)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]any{
			"data": map[string]any{
				"type": "profile-suppression-bulk-create-job",
				"attributes": map[string]any{
					"profiles": map[string]any{
						"data": []map[string]any{
							{"type": "profile", "id": args[0]},
						},
					},
				},
			},
		}

		resp, err := c.Post(ctx, "profile-suppression-bulk-create-jobs", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Profile suppressed.")
		return printData("profiles.suppress", client.FlattenResponse(resp, rawFlag))
	},
}

var profilesMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge two profiles",
	Long:  `Merge two profiles into one. WARNING: This operation cannot be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		source, _ := cmd.Flags().GetString("source")
		destination, _ := cmd.Flags().GetString("destination")
		if source == "" || destination == "" {
			return fmt.Errorf("both --source and --destination profile IDs are required")
		}

		fmt.Fprintf(os.Stderr, "WARNING: Merging profile %s into %s. This cannot be undone.\n", source, destination)

		body := map[string]any{
			"data": map[string]any{
				"type": "profile-merge",
				"relationships": map[string]any{
					"profiles": map[string]any{
						"data": []map[string]any{
							{"type": "profile", "id": source},
							{"type": "profile", "id": destination},
						},
					},
				},
			},
		}

		resp, err := c.Post(ctx, "profile-merge", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Profiles merged.")
		return printData("profiles.merge", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	profilesListCmd.Flags().String("email", "", "Filter by email")
	profilesListCmd.Flags().String("phone", "", "Filter by phone number")

	profilesCreateCmd.Flags().String("email", "", "Email address (required)")
	profilesCreateCmd.Flags().String("first-name", "", "First name")
	profilesCreateCmd.Flags().String("last-name", "", "Last name")
	profilesCreateCmd.Flags().String("phone", "", "Phone number")
	profilesCreateCmd.Flags().String("properties", "", "Custom properties as JSON string")

	profilesUpdateCmd.Flags().String("first-name", "", "First name")
	profilesUpdateCmd.Flags().String("last-name", "", "Last name")
	profilesUpdateCmd.Flags().String("phone", "", "Phone number")
	profilesUpdateCmd.Flags().String("properties", "", "Custom properties as JSON string")

	profilesMergeCmd.Flags().String("source", "", "Source profile ID to merge from (required)")
	profilesMergeCmd.Flags().String("destination", "", "Destination profile ID to merge into (required)")

	profilesCmd.AddCommand(profilesListCmd, profilesGetCmd, profilesCreateCmd, profilesUpdateCmd, profilesSuppressCmd, profilesMergeCmd)
	rootCmd.AddCommand(profilesCmd)
}
