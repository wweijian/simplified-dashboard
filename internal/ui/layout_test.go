package ui

import "testing"

func TestPanelWidths(t *testing.T) {
	tests := []struct {
		total		int
		wantLeft	int
		wantRight	int
	}{
		{100, 40, 59},
		{120, 48, 71},
		{80, 32, 47},
	}
	for _, tt := range tests {
		left, right := panelWidths(tt.total)
		if left != tt.wantLeft || right != tt.wantRight {
			t.Errorf("panelWidths(%d) = (%d, %d), want (%d, %d)",
				tt.total, left, right, tt.wantLeft, tt.wantRight)
		}
	}
}
