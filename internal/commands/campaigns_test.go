package commands

import "testing"

func TestCampaignOmnichannel(t *testing.T) {
	old := revisionFlag
	defer func() { revisionFlag = old }()
	for rev, want := range map[string]bool{"": true, "2026-10-15": true, "2026-12-15": true, "2025-10-15": false, "2024-10-15": false} {
		revisionFlag = rev
		if got := campaignOmnichannel(); got != want {
			t.Errorf("rev %q: got %v want %v", rev, got, want)
		}
	}
}

func TestNormalizeCampaignStatus(t *testing.T) {
	for in, want := range map[string]string{"sent": "Sent", "Sent": "Sent", "SCHEDULED": "Scheduled", "cancelled - no recipients": "Cancelled - No Recipients", "": ""} {
		if got := normalizeCampaignStatus(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}
