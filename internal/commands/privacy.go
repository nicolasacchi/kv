package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var privacyCmd = &cobra.Command{
	Use:   "privacy",
	Short: "GDPR data privacy operations",
	Long: `Request profile deletions and check deletion status.

Examples:
  kv privacy request-deletion --email user@example.com
  kv privacy status <REQUEST_ID>`,
}

var privacyRequestDeletionCmd = &cobra.Command{
	Use:   "request-deletion",
	Short: "Request profile deletion (GDPR)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("email")
		phone, _ := cmd.Flags().GetString("phone")

		if email == "" && phone == "" {
			return fmt.Errorf("either --email or --phone is required")
		}

		profile := map[string]any{}
		if email != "" {
			profile["email"] = email
		}
		if phone != "" {
			profile["phone_number"] = phone
		}

		body := map[string]any{
			"data": map[string]any{
				"type": "data-privacy-deletion-job",
				"attributes": map[string]any{
					"profile": map[string]any{
						"data": map[string]any{
							"type":       "profile",
							"attributes": profile,
						},
					},
				},
			},
		}

		resp, err := c.Post(ctx, "data-privacy-deletion-jobs", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Deletion request submitted.")
		return printData("privacy.request-deletion", client.FlattenResponse(resp, rawFlag))
	},
}

var privacyStatusCmd = &cobra.Command{
	Use:   "status <request-id>",
	Short: "Check deletion request status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "data-privacy-deletion-jobs/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("privacy.status", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	privacyRequestDeletionCmd.Flags().String("email", "", "Email address for deletion request")
	privacyRequestDeletionCmd.Flags().String("phone", "", "Phone number for deletion request")

	privacyCmd.AddCommand(privacyRequestDeletionCmd, privacyStatusCmd)
	rootCmd.AddCommand(privacyCmd)
}
