package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
)

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Manage images",
	Long: `List, get, upload, and update images.

Examples:
  kv images list
  kv images get <ID>
  kv images upload --url https://example.com/image.jpg
  kv images update <ID> --name "New Name"`,
}

var imagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all images",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if noPaginateFlag {
			resp, err := c.Get(ctx, "images", nil)
			if err != nil {
				return err
			}
			return printData("images.list", client.FlattenResponse(resp, rawFlag))
		}

		data, err := c.GetAll(ctx, "images", nil, maxResultsFlag)
		if err != nil {
			return err
		}
		return printData("images.list", client.FlattenRaw(data, rawFlag))
	},
}

var imagesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get image details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "images/"+args[0], nil)
		if err != nil {
			return err
		}
		return printData("images.get", client.FlattenResponse(resp, rawFlag))
	},
}

var imagesUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an image from a URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		urlFlag, _ := cmd.Flags().GetString("url")
		if urlFlag == "" {
			return fmt.Errorf("--url is required")
		}

		attrs := map[string]any{
			"import_from_url": urlFlag,
		}

		name, _ := cmd.Flags().GetString("name")
		if name != "" {
			attrs["name"] = name
		}

		body := jsonapiBody("image", attrs)
		resp, err := c.Post(ctx, "images", body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Image uploaded.")
		return printData("images.upload", client.FlattenResponse(resp, rawFlag))
	},
}

var imagesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an image",
	Args:  cobra.ExactArgs(1),
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

		body := jsonapiBodyWithID("image", args[0], map[string]any{
			"name": name,
		})

		resp, err := c.Patch(ctx, "images/"+args[0], body)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Image updated.")
		return printData("images.update", client.FlattenResponse(resp, rawFlag))
	},
}

func init() {
	imagesUploadCmd.Flags().String("url", "", "Image URL to import (required)")
	imagesUploadCmd.Flags().String("name", "", "Image name")

	imagesUpdateCmd.Flags().String("name", "", "New image name (required)")

	imagesCmd.AddCommand(imagesListCmd, imagesGetCmd, imagesUploadCmd, imagesUpdateCmd)
	rootCmd.AddCommand(imagesCmd)
}
