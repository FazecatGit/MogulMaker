package types

type Bar struct {
	Timestamp string  `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    int64   `json:"v"`
}

type CryptoBar struct {
	Timestamp string  `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    float64 `json:"v"` // Crypto returns volume as float
}

type Candidate struct {
	Symbol         string
	Score          float64
	RSI            float64
	ATR            float64
	Analysis       string
	BodyUpperRatio float64
	BodyLowerRatio float64
	VWAPPrice      float64
	WhaleCount     int
	Bars           []Bar
}

type ScoringInput struct {
	CurrentPrice       float64
	VWAPPrice          float64
	ATRValue           float64
	RSIValue           float64
	WhaleCount         float64
	PriceDrop          float64
	ATRCategory        string
	VolumeRatio        float64 // Current volume / Average volume
	NewsSentimentScore float64 // 0-10 sentiment score
	ShortSignalActive  bool    // Whether short signals are enabled
}
