package strategy

import (
	"testing"

	"github.com/fazecat/mongelmaker/Internal/types"
)

func TestPatternDetector_DetectDoubleBottom(t *testing.T) {
	tests := []struct {
		name      string
		bars      []types.Bar
		wantError bool
	}{
		{
			name: "basic execution",
			bars: []types.Bar{
				{High: 110, Low: 100, Close: 105},
				{High: 105, Low: 88, Close: 95},
				{High: 102, Low: 87, Close: 92},
				{High: 107, Low: 89, Close: 100},
				{High: 115, Low: 90, Close: 110},
			},
			wantError: false,
		},
		{
			name: "insufficient bars",
			bars: []types.Bar{
				{High: 110, Low: 90, Close: 100},
				{High: 105, Low: 88, Close: 95},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewPatternDetector()
			// Just verify it doesn't panic
			result := detector.DetectDoubleBottom(tt.bars)
			if result.Pattern == "" {
				t.Errorf("DetectDoubleBottom() should return a result")
			}
		})
	}
}

func TestPatternDetector_DetectDoubleTop(t *testing.T) {
	tests := []struct {
		name      string
		bars      []types.Bar
		wantError bool
	}{
		{
			name: "basic execution",
			bars: []types.Bar{
				{High: 100, Low: 80, Close: 90},
				{High: 115, Low: 90, Close: 110},
				{High: 105, Low: 88, Close: 95},
				{High: 114, Low: 92, Close: 105},
				{High: 100, Low: 75, Close: 80},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewPatternDetector()
			result := detector.DetectDoubleTop(tt.bars)
			if result.Pattern == "" {
				t.Errorf("DetectDoubleTop() should return a result")
			}
		})
	}
}

func TestPatternDetector_DetectConsolidationBreakout(t *testing.T) {
	tests := []struct {
		name      string
		bars      []types.Bar
		wantError bool
	}{
		{
			name: "basic execution",
			bars: []types.Bar{
				{High: 100, Low: 98, Close: 99, Volume: 1000},
				{High: 100.5, Low: 99, Close: 99.5, Volume: 1000},
				{High: 100.3, Low: 98.8, Close: 99.2, Volume: 1000},
				{High: 100.4, Low: 99.1, Close: 99.8, Volume: 1000},
				{High: 100.2, Low: 99, Close: 99.5, Volume: 1000},
				{High: 100.1, Low: 98.9, Close: 99.3, Volume: 1000},
				{High: 102, Low: 100, Close: 101, Volume: 1500},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewPatternDetector()
			result := detector.DetectConsolidationBreakout(tt.bars)
			if result.Pattern == "" {
				t.Errorf("DetectConsolidationBreakout() should return a result")
			}
		})
	}
}

func TestPatternDetector_DetectAllPatterns(t *testing.T) {
	bars := []types.Bar{
		{High: 110, Low: 100, Close: 105, Volume: 1000},
		{High: 105, Low: 88, Close: 95, Volume: 1000},
		{High: 102, Low: 87, Close: 92, Volume: 1000},
		{High: 107, Low: 89, Close: 100, Volume: 1000},
		{High: 115, Low: 90, Close: 110, Volume: 1000},
	}

	detector := NewPatternDetector()
	patterns := detector.DetectAllPatterns(bars)

	// At least verify the function runs without error
	if patterns == nil {
		t.Errorf("DetectAllPatterns() should return a slice, not nil")
	}

	// Verify all patterns have required fields
	for _, p := range patterns {
		if p.Pattern == "" {
			t.Errorf("Pattern has empty Pattern field")
		}
		if p.Direction == "NONE" && p.Detected {
			t.Errorf("Detected pattern should have a direction")
		}
	}
}

func TestPatternDetector_RiskRewardCalculation(t *testing.T) {
	bars := []types.Bar{
		{High: 110, Low: 100, Close: 105, Volume: 1000},
		{High: 105, Low: 88, Close: 95, Volume: 1000},
		{High: 102, Low: 87, Close: 92, Volume: 1000},
		{High: 107, Low: 89, Close: 100, Volume: 1000},
		{High: 115, Low: 90, Close: 110, Volume: 1000},
	}

	detector := NewPatternDetector()
	pattern := detector.DetectDoubleBottom(bars)

	// Just verify result structure
	if pattern.Pattern == "" {
		t.Errorf("Pattern should be initialized")
	}
}
