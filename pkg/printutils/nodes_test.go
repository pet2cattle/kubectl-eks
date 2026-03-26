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
