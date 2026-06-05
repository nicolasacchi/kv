package commands

import (
	"fmt"
	"os"

	"github.com/nicolasacchi/kv/internal/client"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Write-safety flags. Destructive Klaviyo verbs (resource deletes, profile
// suppress/merge, GDPR profile deletion) refuse unless the operator opts in.
// Registered as persistent flags on rootCmd.
var (
	yesFlag    bool
	dryRunFlag bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&yesFlag, "yes", false, "Confirm destructive operations (alias: --confirm)")
	rootCmd.PersistentFlags().BoolVar(&yesFlag, "confirm", false, "Alias for --yes")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Print the intended mutation and exit without sending")
}

// requireConfirm gates a state-changing verb. It returns a write_locked APIError
// (exit 6) when --yes/--confirm was not passed, so callers and agents can
// dispatch on Kind and tell "refused" apart from a real API failure.
func requireConfirm(action string) error {
	if yesFlag {
		return nil
	}
	hint := "re-run with --yes (or --confirm) to proceed"
	if term.IsTerminal(int(os.Stdout.Fd())) {
		hint = "re-run with --yes to proceed, or --dry-run to preview"
	}
	return &client.APIError{
		Kind:   "write_locked",
		Detail: fmt.Sprintf("%s requires confirmation", action),
		Hint:   hint,
	}
}

// gate guards a destructive verb in one line at the top of a RunE:
//
//	if handled, err := gate(cmd, fmt.Sprintf("delete list %s", id)); handled {
//	    return err
//	}
//
// It returns handled=true when the caller should return immediately: either
// --dry-run previewed the action (err nil, exit 0) or the gate refused
// (err is a write_locked APIError, exit 6).
func gate(cmd *cobra.Command, action string) (bool, error) {
	if dryRunFlag {
		fmt.Fprintf(cmd.OutOrStdout(), "--dry-run: would %s, no changes made\n", action)
		return true, nil
	}
	if err := requireConfirm(action); err != nil {
		return true, err
	}
	return false, nil
}
