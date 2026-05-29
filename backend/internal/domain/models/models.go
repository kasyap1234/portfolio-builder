package models

import (
	"time"
)

type AssetType string

const (
	AssetTypeStock AssetType = "STOCK"
	AssetTypeMF    AssetType = "MF"
	AssetTypeIndex AssetType = "INDEX" // For Nifty 50
	AssetTypeETF   AssetType = "ETF"   // For Nifty50 ETF
	AssetTypeDebt  AssetType = "DEBT"
)

type Asset struct {
	Symbol   string
	Type     AssetType
	Quantity float64
	AvgPrice float64
}

type Portfolio struct {
	Assets []Asset
	Cash   float64
}

// AllocationSource distinguishes monthly SIP budget from debt-reserve deployments.
type AllocationSource string

const (
	AllocationSourceSIP         AllocationSource = "sip"
	AllocationSourceDebtReserve AllocationSource = "debt_reserve"
)

type AllocationRecommendation struct {
	AssetSymbol string
	AssetType   AssetType
	Amount      float64
	Reason      string
	Source      AllocationSource // empty defaults to SIP
}

type MarketStatus struct {
	NiftyPE          float64
	NiftyPrice       float64
	NiftyDMA200      float64
	NiftyDropMetrics *DropMetrics
}

// DropMetrics captures various drop/discount metrics for an asset
type DropMetrics struct {
	Symbol          string
	CurrentPrice    float64
	WeekAgoPrice    float64 // Price 1 week ago
	MonthAgoPrice   float64 // Price 1 month ago
	RecentHighPrice float64 // 52-week or 1-year high
	DMA200          float64 // 200-day moving average
	WeekDrop        float64 // % drop in last week (positive = down)
	MonthDrop       float64 // % drop in last month (positive = down)
	HighDrop        float64 // % drop from recent high (positive = down)
	DMADistance     float64 // % distance from 200 DMA (positive = below DMA)
	CompositeScore  float64 // Weighted score based on all metrics
}

// Weights for composite score calculation
const (
	WeightDMA200     = 0.40
	WeightRecentHigh = 0.30
	WeightWeekDrop   = 0.15
	WeightMonthDrop  = 0.15
)

// CalculateCompositeScore calculates the weighted composite "drop score"
// Higher score = more "down" = more attractive for buying
func (d *DropMetrics) CalculateCompositeScore() float64 {
	d.CompositeScore = (d.DMADistance * WeightDMA200) +
		(d.HighDrop * WeightRecentHigh) +
		(d.WeekDrop * WeightWeekDrop) +
		(d.MonthDrop * WeightMonthDrop)
	return d.CompositeScore
}

// IsMoreDownThan returns true if this asset is more "down" than the reference asset
func (d *DropMetrics) IsMoreDownThan(ref *DropMetrics) bool {
	return d.CompositeScore > ref.CompositeScore
}

// TradeLog represents a single transaction record
type TradeLog struct {
	ID              int64
	AssetID         int64
	TransactionDate time.Time
	TransactionType string  // BUY, SELL, DIVIDEND
	Amount          float64 // INR amount
	Units           float64 // Units purchased
	PricePerUnit    float64
	AssetCategory   string // EQUITY, DEBT, ETF, MF
	Notes           string
	CreatedAt       time.Time
}

// AllocationSnapshot captures portfolio state at a point in time
type AllocationSnapshot struct {
	ID                  int64
	SnapshotDate        time.Time
	TotalEquityAmount   float64
	TotalDebtAmount     float64
	EquityAllocationPct float64
	DebtAllocationPct   float64
	MarketRegime        string
	NiftyDMADistance    float64
	CreatedAt           time.Time
}

// DebtReserve tracks additional cash for panic buying deployment
type DebtReserve struct {
	ID                 int64
	TotalAmount        float64
	DeployedAmount     float64
	AvailableAmount    float64
	LastDeploymentDate *time.Time
	UpdatedAt          time.Time
}

// Transaction type constants
const (
	TransactionTypeBuy      = "BUY"
	TransactionTypeSell     = "SELL"
	TransactionTypeDividend = "DIVIDEND"
)

// Asset category constants
const (
	CategoryEquity = "EQUITY"
	CategoryDebt   = "DEBT"
	CategoryETF    = "ETF"
	CategoryMF     = "MF"
)
