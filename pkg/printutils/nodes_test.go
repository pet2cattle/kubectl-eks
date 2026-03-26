package printutils

import (
	"testing"

	"github.com/jordiprats/kubectl-eks/pkg/data"
)

func TestFormatNodeConditions(t *testing.T) {
	tests := []struct {
		name string
		node data.NodeInfo
		want string
	}{
		{
			name: "no active conditions",
			node: data.NodeInfo{
				MemoryPressure:     "False",
				DiskPressure:       "Unknown",
				PIDPressure:        "False",
				NetworkUnavailable: "Unknown",
			},
			want: "-",
		},
		{
			name: "single active condition",
			node: data.NodeInfo{
				DiskPressure: "True",
			},
			want: "DiskPressure",
		},
		{
			name: "multiple active conditions",
			node: data.NodeInfo{
				MemoryPressure:     "True",
				PIDPressure:        "True",
				NetworkUnavailable: "True",
			},
			want: "MemoryPressure,PIDPressure,NetworkUnavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatNodeConditions(tt.node)
			if got != tt.want {
				t.Fatalf("formatNodeConditions() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatUsedTotal(t *testing.T) {
	if got := formatUsedTotal("500m", "4"); got != "500m/4" {
		t.Fatalf("formatUsedTotal() = %q, want %q", got, "500m/4")
	}
	if got := formatUsedTotal("-", "-"); got != "-/-" {
		t.Fatalf("formatUsedTotal() = %q, want %q", got, "-/-")
	}
}

func TestFormatUsedTotalRemaining(t *testing.T) {
	if got := formatUsedTotalRemaining("500m", "4", "3500m"); got != "500m/4 (3500m)" {
		t.Fatalf("formatUsedTotalRemaining() = %q, want %q", got, "500m/4 (3500m)")
	}
	if got := formatUsedTotalRemaining("-", "-", "-"); got != "-/- (-)" {
		t.Fatalf("formatUsedTotalRemaining() = %q, want %q", got, "-/- (-)")
	}
}

func TestFormatCPUQuantityCores(t *testing.T) {
	if got := formatCPUQuantityCores("500m"); got != "0.5" {
		t.Fatalf("formatCPUQuantityCores() = %q, want %q", got, "0.5")
	}
	if got := formatCPUQuantityCores("4"); got != "4" {
		t.Fatalf("formatCPUQuantityCores() = %q, want %q", got, "4")
	}
	if got := formatCPUQuantityCores("3500m"); got != "3.5" {
		t.Fatalf("formatCPUQuantityCores() = %q, want %q", got, "3.5")
	}
	if got := formatCPUQuantityCores("-"); got != "-" {
		t.Fatalf("formatCPUQuantityCores() = %q, want %q", got, "-")
	}
}

func TestFormatCPUUsedTotalRemaining(t *testing.T) {
	if got := formatCPUUsedTotalRemaining("500m", "4", "3500m"); got != "0.5/4 (3.5)" {
		t.Fatalf("formatCPUUsedTotalRemaining() = %q, want %q", got, "0.5/4 (3.5)")
	}
	if got := formatCPUUsedTotalRemaining("-", "-", "-"); got != "-/- (-)" {
		t.Fatalf("formatCPUUsedTotalRemaining() = %q, want %q", got, "-/- (-)")
	}
}

func TestFormatMemoryQuantityGi(t *testing.T) {
	if got := formatMemoryQuantityGi("15984776Ki"); got != "15.2Gi" {
		t.Fatalf("formatMemoryQuantityGi() = %q, want %q", got, "15.2Gi")
	}
	if got := formatMemoryQuantityGi("2248Mi"); got != "2.2Gi" {
		t.Fatalf("formatMemoryQuantityGi() = %q, want %q", got, "2.2Gi")
	}
	if got := formatMemoryQuantityGi("14Gi"); got != "14Gi" {
		t.Fatalf("formatMemoryQuantityGi() = %q, want %q", got, "14Gi")
	}
	if got := formatMemoryQuantityGi("-"); got != "-" {
		t.Fatalf("formatMemoryQuantityGi() = %q, want %q", got, "-")
	}
}

func TestFormatMemoryUsedTotalRemaining(t *testing.T) {
	if got := formatMemoryUsedTotalRemaining("2248Mi", "15984776Ki", "13682824Ki"); got != "2.2Gi/15.2Gi (13Gi)" {
		t.Fatalf("formatMemoryUsedTotalRemaining() = %q, want %q", got, "2.2Gi/15.2Gi (13Gi)")
	}
	if got := formatMemoryUsedTotalRemaining("-", "-", "-"); got != "-/- (-)" {
		t.Fatalf("formatMemoryUsedTotalRemaining() = %q, want %q", got, "-/- (-)")
	}
}
