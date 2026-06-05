package commands

import (
	"errors"
	"testing"

	"github.com/nicolasacchi/kv/internal/client"
)

func TestRequireConfirm(t *testing.T) {
	cases := []struct {
		name    string
		yes     bool
		wantErr bool
	}{
		{"refuses without --yes", false, true},
		{"proceeds with --yes", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orig := yesFlag
			t.Cleanup(func() { yesFlag = orig })
			yesFlag = tc.yes

			err := requireConfirm("delete list abc")
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("expected nil with --yes, got %v", err)
				}
				return
			}
			var apiErr *client.APIError
			if !errors.As(err, &apiErr) || apiErr.Kind != "write_locked" {
				t.Fatalf("expected write_locked APIError, got %v", err)
			}
			if apiErr.ExitCode() != 6 {
				t.Fatalf("write_locked must map to exit 6, got %d", apiErr.ExitCode())
			}
		})
	}
}

// TestGate covers the one-line helper: --dry-run previews (handled, nil) and
// the default refusal (handled, write_locked).
func TestGate(t *testing.T) {
	cmd := rootCmd

	t.Run("dry-run previews and is handled", func(t *testing.T) {
		origDry, origYes := dryRunFlag, yesFlag
		t.Cleanup(func() { dryRunFlag, yesFlag = origDry, origYes })
		dryRunFlag, yesFlag = true, false
		handled, err := gate(cmd, "delete list abc")
		if !handled || err != nil {
			t.Fatalf("dry-run: want handled=true err=nil, got handled=%v err=%v", handled, err)
		}
	})

	t.Run("refuses without --yes and is handled", func(t *testing.T) {
		origDry, origYes := dryRunFlag, yesFlag
		t.Cleanup(func() { dryRunFlag, yesFlag = origDry, origYes })
		dryRunFlag, yesFlag = false, false
		handled, err := gate(cmd, "delete list abc")
		if !handled {
			t.Fatal("refusal must be handled")
		}
		var apiErr *client.APIError
		if !errors.As(err, &apiErr) || apiErr.ExitCode() != 6 {
			t.Fatalf("want write_locked exit 6, got %v", err)
		}
	})

	t.Run("proceeds with --yes (not handled)", func(t *testing.T) {
		origDry, origYes := dryRunFlag, yesFlag
		t.Cleanup(func() { dryRunFlag, yesFlag = origDry, origYes })
		dryRunFlag, yesFlag = false, true
		handled, err := gate(cmd, "delete list abc")
		if handled || err != nil {
			t.Fatalf("with --yes: want handled=false err=nil, got handled=%v err=%v", handled, err)
		}
	})
}
