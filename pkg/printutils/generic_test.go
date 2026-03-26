package printutils

import "testing"

func TestFormatClusterNodeHealth(t *testing.T) {
	if got := formatClusterNodeHealth(0, 0, 0, 0); got != "-" {
		t.Fatalf("formatClusterNodeHealth() = %q, want %q", got, "-")
	}

	if got := formatClusterNodeHealth(10, 9, 1, 0); got != "9/10 Ready (NR:1 SD:0)" {
		t.Fatalf("formatClusterNodeHealth() = %q, want %q", got, "9/10 Ready (NR:1 SD:0)")
	}
}
