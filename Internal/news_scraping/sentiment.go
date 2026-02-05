package newsscraping

import "strings"

type SentimentAnalyzer struct {
	positiveWords map[string]float64
	negativeWords map[string]float64
}

func NewSentimentAnalyzer() *SentimentAnalyzer {
	return &SentimentAnalyzer{
		positiveWords: map[string]float64{
			// Strong positive (0.9-1.0)
			"surge": 1.0, "soar": 1.0, "skyrocket": 1.0, "breakthrough": 1.0,
			"bullish": 0.95, "rally": 0.95, "boom": 0.95, "explode": 0.95,
			"rocket": 0.9, "triumph": 0.9, "outperform": 0.9, "breakout": 0.9,

			// Moderate positive (0.7-0.89)
			"beat": 0.85, "exceed": 0.85, "upgrade": 0.85, "optimistic": 0.85,
			"profit": 0.8, "growth": 0.8, "gain": 0.8, "jump": 0.8,
			"strong": 0.8, "boost": 0.8, "success": 0.8, "win": 0.8,
			"improve": 0.75, "rising": 0.75, "advance": 0.75, "climb": 0.75,
			"expansion": 0.75, "momentum": 0.75, "upside": 0.75, "favorable": 0.75,
			"recover": 0.7, "rebound": 0.7, "stabilize": 0.7, "strength": 0.7,

			// Mild positive (0.5-0.69)
			"positive": 0.65, "rise": 0.65, "higher": 0.65, "increase": 0.65,
			"better": 0.65, "good": 0.65, "solid": 0.65, "confident": 0.65,
			"opportunity": 0.6, "potential": 0.6, "promising": 0.6, "attractive": 0.6,
			"value": 0.6, "support": 0.6, "resilient": 0.6, "steady": 0.6,
			"healthy": 0.55, "buying": 0.55, "progress": 0.55, "achievement": 0.55,
			"innovative": 0.55, "leader": 0.55, "advantage": 0.55, "superior": 0.55,
			"efficient": 0.5, "robust": 0.5, "stable": 0.5, "quality": 0.5,
		},
		negativeWords: map[string]float64{
			// Strong negative (0.9-1.0)
			"crash": 1.0, "plunge": 1.0, "collapse": 1.0, "devastate": 1.0,
			"catastrophic": 1.0, "disaster": 1.0, "crisis": 0.95, "bankruptcy": 0.95,
			"savaged": 0.95, "plummet": 0.95, "tumble": 0.95, "rout": 0.95,
			"hammered": 0.9, "slaughter": 0.9, "massacre": 0.9, "panic": 0.9,

			// Moderate negative (0.7-0.89)
			"bearish": 0.85, "downgrade": 0.85, "warning": 0.85, "alert": 0.85,
			"lawsuit": 0.85, "lawsuits": 0.85, "class action": 0.85, "delinquency": 0.85,
			"delinquencies": 0.85, "scrutiny": 0.85, "dispute": 0.8, "disputes": 0.8,
			"miss": 0.8, "loss": 0.8, "losses": 0.8, "slump": 0.8,
			"decline": 0.8, "deteriorate": 0.8, "underperform": 0.8, "fail": 0.8,
			"struggle": 0.75, "struggles": 0.75, "weak": 0.75, "weakness": 0.75,
			"drop": 0.75, "fall": 0.75, "falls": 0.75, "falling": 0.75,
			"concern": 0.7, "concerns": 0.7, "worry": 0.7, "worries": 0.7,
			"unhappy": 0.7, "disappoint": 0.7, "disappoints": 0.7, "challenges": 0.7,
			"uncertain": 0.7, "risky": 0.7, "expose": 0.7,

			// Mild negative (0.5-0.69)
			"problem": 0.65, "problems": 0.65, "issue": 0.65, "issues": 0.65,
			"risk": 0.65, "risks": 0.65, "threat": 0.65, "volatile": 0.65,
			"uncertainty": 0.65, "unclear": 0.65, "doubt": 0.65, "question": 0.65,
			"pressure": 0.6, "challenge": 0.6, "difficult": 0.6, "hurt": 0.6,
			"lower": 0.6, "below": 0.6, "under": 0.6, "disappointing": 0.6,
			"negative": 0.6, "poor": 0.6, "slow": 0.6, "slowdown": 0.6,
			"dip": 0.55, "slip": 0.55, "soften": 0.55, "retreat": 0.55,
			"caution": 0.55, "cautious": 0.55, "contrarian": 0.55, "downside": 0.55,
			"correction": 0.5, "pullback": 0.5, "trim": 0.5, "cut": 0.5,
			"reduce": 0.5, "drag": 0.5, "obstacle": 0.5, "headwind": 0.5,
			"bigger": 0.5, "worst": 0.9,
		},
	}
}

func (sa *SentimentAnalyzer) Analyze(text string) (SentimentScore, float64) {
	text = strings.ToLower(text)
	words := strings.Fields(text)

	var score float64
	var matches int

	for _, word := range words {
		word = strings.Trim(word, ".,!?\"'()[]{}:;")

		if val, exists := sa.positiveWords[word]; exists {
			score += val
			matches++
		} else if val, exists := sa.negativeWords[word]; exists {
			score -= val
			matches++
		}
	}

	if matches > 0 {
		score /= float64(matches)
	}
	sentiment := Neutral
	if score > 0.1 {
		sentiment = Positive
	} else if score < -0.1 {
		sentiment = Negative
	}
	return sentiment, score

}
