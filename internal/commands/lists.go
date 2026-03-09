package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage lists",
	Long: `List, get, create lists and manage list members.

Examples:
  kv lists list
  kv lists get <ID>
  kv lists create --name "My List"
  kv lists members <ID>
  kv lists add-member <LIST_ID> --profile <PROFILE_ID>`,
}

var listsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "lists", nil)
			if err != nil {
				return err
			}
			return printData("lists.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "lists", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("lists.list", client.FlattenRaw(data, rawFlag))
	},
}

var listsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get list details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "lists/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("lists.get", client.FlattenResponse(resp, rawFlag))
	},
}

var listsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new list",
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

		body := jsonapiBody("list", map[string]any{
			"name": name,
		})

		resp, err := c.Post(ctx, "lists", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "List created.")
		return printData("lists.create", client.FlattenResponse(resp, rawFlag))
	},
}

var listsMembersCmd = &cobra.Command{
	Use:   "members <list-id>",
	Short: "List profiles in a list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("lists/%s/profiles", args[0])
		if noPaginateFlag {
			resp, err := c.Get(ctx, path, nil)
			if err != nil {
				return err
			}
			return printData("lists.members", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, path, nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("lists.members", client.FlattenRaw(data, rawFlag))
	},
}

var listsAddMemberCmd = &cobra.Command{
	Use:   "add-member <list-id>",
	Short: "Add a profile to a list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		profileID, _ := cmd.Flags().GetString("profile")
		if profileID == "" {
			return fmt.Errorf("--profile is required")
		}

		body := map[string]any{
			"data": []map[string]any{
				{"type": "profile", "id": profileID},
			},
		}

		path := fmt.Sprintf("lists/%s/relationships/profiles", args[0])
		_, err = c.Post(ctx, path, body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Profile added to list.")
		return nil
	},
}

var listsRemoveMemberCmd = &cobra.Command{
	Use:   "remove-member <list-id>",
	Short: "Remove a profile from a list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		profileID, _ := cmd.Flags().GetString("profile")
		if profileID == "" {
			return fmt.Errorf("--profile is required")
		}

		path := fmt.Sprintf("lists/%s/relationships/profiles", args[0])
		err = c.Delete(ctx, path)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Profile removed from list.")
		return nil
	},
}

func init() {
	listsCreateCmd.Flags().String("name", "", "List name (required)")
	listsAddMemberCmd.Flags().String("profile", "", "Profile ID to add (required)")
	listsRemoveMemberCmd.Flags().String("profile", "", "Profile ID to remove (required)")

	listsCmd.AddCommand(listsListCmd, listsGetCmd, listsCreateCmd, listsMembersCmd, listsAddMemberCmd, listsRemoveMemberCmd)
	rootCmd.AddCommand(listsCmd)
}
