package detection

import (
	"testing"

	"github.com/fazecat/mongelmaker/Internal/types"
	"github.com/fazecat/mongelmaker/Internal/utils"
	"github.com/fazecat/mongelmaker/Internal/utils/config"
)

func TestCalculateInterestScore_NeutralConditions(t *testing.T) {
	weights := config.SignalWeights{
		RSIWeight:           0.2,
		ATRWeight:           0.2,
		VolumeWeight:        0.2,
		NewsSentimentWeight: 0.2,
		WhaleActivityWeight: 0.2,
	}
	input := types.ScoringInput{
		CurrentPrice: 100,
		VWAPPrice:    100,
		ATRValue:     1.0,
		RSIValue:     50,
		WhaleCount:   0,
		PriceDrop:    0,
		ATRCategory:  "NORMAL",
	}

	score := CalculateInterestScore(input, weights)
	expected := 5.0
	tolerance := 0.15
	if utils.Abs(score-expected) > tolerance {
		t.Errorf("Expected score %f (±%f), got %f", expected, tolerance, score)
	}
}

func TestCalculateInterestScore_PriceDropBoost(t *testing.T) {
	weights := config.SignalWeights{
		RSIWeight:           0.2,
		ATRWeight:           0.2,
		VolumeWeight:        0.2,
		NewsSentimentWeight: 0.2,
		WhaleActivityWeight: 0.2,
	}
	input := types.ScoringInput{
		CurrentPrice: 80,
		VWAPPrice:    100,
		ATRValue:     1.0,
		RSIValue:     50,
		WhaleCount:   0,
		PriceDrop:    20,
		ATRCategory:  "NORMAL",
	}
	score := CalculateInterestScore(input, weights)
	expectedMin := 5.0 // Adjust as needed for your logic
	if score < expectedMin {
		t.Errorf("Expected score at least %f, got %f", expectedMin, score)
	}
}

func TestCalculateInterestScore_HighWhaleActivity(t *testing.T) {
	weights := config.SignalWeights{
		RSIWeight:           0.2,
		ATRWeight:           0.2,
		VolumeWeight:        0.2,
		NewsSentimentWeight: 0.2,
		WhaleActivityWeight: 0.2,
	}
	input := types.ScoringInput{
		CurrentPrice: 100,
		VWAPPrice:    100,
		ATRValue:     1.0,
		RSIValue:     50,
		WhaleCount:   10,
		PriceDrop:    0,
		ATRCategory:  "NORMAL",
	}
	score := CalculateInterestScore(input, weights)
	expectedMin := 5.0 * 1.15
	if score < expectedMin {
		t.Errorf("Expected score at least %f, got %f", expectedMin, score)
	}
}

func TestCalculateInterestScore_LowRSIPenalty(t *testing.T) {
	weights := config.SignalWeights{
		RSIWeight:           0.2,
		ATRWeight:           0.2,
		VolumeWeight:        0.2,
		NewsSentimentWeight: 0.2,
		WhaleActivityWeight: 0.2,
	}
	input := types.ScoringInput{
		CurrentPrice: 100,
		VWAPPrice:    100,
		ATRValue:     1.0,
		RSIValue:     25,
		WhaleCount:   0,
		PriceDrop:    0,
		ATRCategory:  "NORMAL",
	}
	score := CalculateInterestScore(input, weights)
	expectedMax := 6.0
	if score > expectedMax {
		t.Errorf("Expected score at most %f, got %f", expectedMax, score)
	}
}

func TestCalculateInterestScore_HighATRAdjustment(t *testing.T) {
	weights := config.SignalWeights{
		RSIWeight:           0.2,
		ATRWeight:           0.2,
		VolumeWeight:        0.2,
		NewsSentimentWeight: 0.2,
		WhaleActivityWeight: 0.2,
	}
	input := types.ScoringInput{
		CurrentPrice: 100,
		VWAPPrice:    100,
		ATRValue:     3.0,
		RSIValue:     50,
		WhaleCount:   0,
		PriceDrop:    0,
		ATRCategory:  "HIGH",
	}
	score := CalculateInterestScore(input, weights)
	expectedMin := 5.0
	if score < expectedMin {
		t.Errorf("Expected score at least %f, got %f", expectedMin, score)
	}
}
