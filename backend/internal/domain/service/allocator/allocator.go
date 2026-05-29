package allocator

import (
	"fmt"
	"math"
	"smart-alert/internal/domain/models"
	service "smart-alert/internal/domain/service/data_fetcher"
	"sort"
	"strings"
	"sync"
)

// Predefined stock watchlist with Yahoo Finance symbols
var StockWatchlist = []string{
	"RELIANCE.NS",
	"CIPLA.NS",
	"HAVELLS.NS",
	"ICICIBANK.NS",
	"SHRIRAMFIN.NS",
	"RAINBOW.NS", // Rainbow Children's Medicare
	"KPITTECH.NS",
	"SENCO.NS",
	"ZAGGLE.NS",
	"CAMS.NS",
	"BAJFINANCE.NS",
}

// Fallback market cap data (in INR) - updated periodically
// Used when API fails to fetch market cap
var FallbackMarketCaps = map[string]float64{
	"RELIANCE.NS":   17000000000000, // ~17 Lakh Cr - Large Cap
	"CIPLA.NS":      1200000000000,  // ~1.2 Lakh Cr - Large Cap
	"HAVELLS.NS":    1000000000000,  // ~1 Lakh Cr - Large Cap
	"ICICIBANK.NS":  9000000000000,  // ~9 Lakh Cr - Large Cap
	"SHRIRAMFIN.NS": 1100000000000,  // ~1.1 Lakh Cr - Large Cap
	"RAINBOW.NS":    150000000000,   // ~15k Cr - Small Cap
	"KPITTECH.NS":   300000000000,   // ~30k Cr - Mid Cap
	"SENCO.NS":      100000000000,   // ~10k Cr - Small Cap
	"ZAGGLE.NS":     50000000000,    // ~5k Cr - Small Cap
	"CAMS.NS":       180000000000,   // ~18k Cr - Small Cap
	"BAJFINANCE.NS": 4500000000000,  // ~4.5 Lakh Cr - Large Cap
}

// Predefined MF watchlist with AMFI codes
var MFWatchlist = []string{
	"150346", // Whiteoak Capital Flexi Cap Fund Direct (G)
}

// Nifty 50 symbol for Yahoo Finance
const NiftySymbol = "^NSEI"

// Nifty50 ETF for panic buying deployment
const Nifty50ETFSymbol = "NIFTY50-ETF" // Nippon India ETF Nifty 50 BeES

// WhiteOak Flexi Cap Fund AMFI code
const WhiteoakFlexiCapCode = "150346"

// Debt reserve deployment percentage during DEEP_FEAR
const DebtReserveDeploymentPct = 0.50 // Deploy 50% of debt reserve in DEEP_FEAR

// Debt reserve deployment percentage during FEAR
const DebtReserveFearDeploymentPct = 0.02 // Deploy 2% of debt reserve in FEAR

// PE based deployment
const DefaultPETriggerThreshold = 19.0 // Trigger bulk buy when PE <= 19
const PEDeploymentPct = 0.30           // Deploy 30% of debt reserve

// ============================================
// Market Regime Based Allocation Constants
// ============================================

// Market Regime Thresholds (based on effective Nifty distance from 200 DMA)
const (
	DeepFearThreshold = 10.0 // Nifty > 10% below 200 DMA (after PE adjustment)
	FearThreshold     = 0.0  // Nifty below 200 DMA (but < 10%)
	NeutralUpperBound = 5.0  // Nifty within ±5% of 200 DMA
	GreedThreshold    = 5.0  // Nifty > 5% above 200 DMA
)

// Nifty PE adjustment for regime detection (weighted more heavily than raw DMA)
const (
	PENeutralReference = 22.0 // Center of fair-value band for Nifty 50
	PEFearBoostWeight  = 2.0  // Each PE point below neutral adds this % to effective DMA distance
)

// Asset Allocation by Market Regime
type MarketRegime string

const (
	RegimeDeepFear MarketRegime = "DEEP_FEAR"
	RegimeFear     MarketRegime = "FEAR"
	RegimeNeutral  MarketRegime = "NEUTRAL"
	RegimeGreed    MarketRegime = "GREED"
)

// Allocation percentages per regime [Equity, Debt]
var RegimeAllocations = map[MarketRegime]struct {
	Equity float64
	Debt   float64
}{
	RegimeDeepFear: {Equity: 1.00, Debt: 0.00}, // Maximum equity in deep fear
	RegimeFear:     {Equity: 0.80, Debt: 0.20},
	RegimeNeutral:  {Equity: 0.60, Debt: 0.40},
	RegimeGreed:    {Equity: 0.30, Debt: 0.70}, // Maximum debt in greed
}

// Stock Selection Thresholds
const (
	StockDropMultiplier     = 2.0  // Stock must fall 2x Nifty to qualify
	SharpWeeklyDrop         = 10.0 // 10% weekly drop is "sharp"
	SharpMonthlyDrop        = 15.0 // 15% monthly drop is "sharp"
	MinStockAllocation      = 0.03 // Min 3% of equity can go to stocks (prefer MF)
	MaxStockAllocation      = 0.30 // Max 30% of equity can go to stocks
	MaxSingleStockPercent   = 0.25 // Max 25% of stock allocation per stock
	MaxSmallCapTotalPercent = 0.25 // Max 25% of stock allocation to small caps combined
)

// Volatility/Risk weights by Market Cap
// Lower weight = higher volatility = less allocation
const (
	LargeCapThreshold = 1000000000000 // > 1 Lakh Crore
	MidCapThreshold   = 300000000000  // > 30k Crore
	LargeCapWeight    = 1.0           // Full weight for stable large caps
	MidCapWeight      = 0.6           // 60% weight for mid caps
	SmallCapWeight    = 0.10          // 10% weight for volatile small caps
	UnknownCapWeight  = 0.35          // 35% weight when market cap unknown (conservative)
)

type Allocator interface {
	Allocate(amount float64, currentPortfolio []models.Asset, cashInHand float64, watchlist []string) ([]models.AllocationRecommendation, error)
	CalculateDropMetrics(symbol string) (*models.DropMetrics, error)
	DetermineMarketRegime(niftyMetrics *models.DropMetrics, niftyPE float64) MarketRegime
}

type allocator struct {
	fetcher            service.DataFetcher
	peTriggerThreshold float64
}

func NewAllocator(fetcher service.DataFetcher, opts ...AllocatorOption) Allocator {
	a := &allocator{
		fetcher:            fetcher,
		peTriggerThreshold: DefaultPETriggerThreshold,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

type AllocatorOption func(*allocator)

func WithPETriggerThreshold(threshold float64) AllocatorOption {
	return func(a *allocator) {
		a.peTriggerThreshold = threshold
	}
}

func (a *allocator) Allocate(amount float64, currentPortfolio []models.Asset, cashInHand float64, watchlist []string) ([]models.AllocationRecommendation, error) {
	// 1. Get Nifty drop metrics (benchmark)
	niftyMetrics, err := a.CalculateDropMetrics(NiftySymbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get Nifty metrics: %v", err)
	}

	// 2. Determine market regime (DMA + Nifty PE; PE weighted heavily)
	niftyPE, _ := a.fetcher.FetchPE(NiftySymbol)
	regime := a.DetermineMarketRegime(niftyMetrics, niftyPE)
	allocation := RegimeAllocations[regime]

	recommendations := []models.AllocationRecommendation{}

	// 3. Calculate base amounts
	equityAmount := amount * allocation.Equity
	debtAmount := amount * allocation.Debt

	// 4. Deploy debt reserve based on market regime
	// DEEP_FEAR: Deploy 50%, FEAR: Deploy 2% (50-50 split: Nifty50 ETF + Whiteoak)
	if cashInHand > 0 {
		remainingReserve := cashInHand
		var deploymentPct float64
		var regimeLabel string

		switch regime {
		case RegimeDeepFear:
			deploymentPct = DebtReserveDeploymentPct
			regimeLabel = "DEEP_FEAR"
		case RegimeFear:
			deploymentPct = DebtReserveFearDeploymentPct
			regimeLabel = "FEAR"
		}

		if deploymentPct > 0 {
			debtDeployment := cashInHand * deploymentPct
			if debtDeployment > remainingReserve {
				debtDeployment = remainingReserve
			}
			halfDeployment := debtDeployment / 2

			// Nifty50 ETF allocation (50% of deployment)
			remainingReserve -= halfDeployment
			recommendations = append(recommendations, models.AllocationRecommendation{
				AssetSymbol: Nifty50ETFSymbol,
				AssetType:   models.AssetTypeETF,
				Amount:      halfDeployment,
				Source:      models.AllocationSourceDebtReserve,
				Reason:      reasonPanicBuyETF(deploymentPct, debtDeployment, halfDeployment, regimeLabel, niftyMetrics.DMADistance),
			})

			// Whiteoak Flexi Cap Fund allocation (50% of deployment)
			remainingReserve -= halfDeployment
			recommendations = append(recommendations, models.AllocationRecommendation{
				AssetSymbol: WhiteoakFlexiCapCode,
				AssetType:   models.AssetTypeMF,
				Amount:      halfDeployment,
				Source:      models.AllocationSourceDebtReserve,
				Reason:      reasonPanicBuyMF(deploymentPct, debtDeployment, halfDeployment, regimeLabel, niftyMetrics.DMADistance),
			})
		}

		// 4.1 Check Nifty PE Trigger (Independent and Additive, bounded by remaining reserve)
		if niftyPE > 0 && niftyPE <= a.peTriggerThreshold {
			peDeployment := cashInHand * PEDeploymentPct
			if peDeployment > remainingReserve {
				peDeployment = remainingReserve
			}
			if peDeployment > 0 {
				remainingReserve -= peDeployment
				recommendations = append(recommendations, models.AllocationRecommendation{
					AssetSymbol: Nifty50ETFSymbol,
					AssetType:   models.AssetTypeETF,
					Amount:      peDeployment,
					Source:      models.AllocationSourceDebtReserve,
					Reason:      reasonPETrigger(niftyPE, a.peTriggerThreshold, PEDeploymentPct, peDeployment),
				})
			}
		}
	}

	// 5. Distribute equity between MF and Stocks
	stockRecs, stockTotal, err := a.distributeToStocks(equityAmount, niftyMetrics, regime)
	if err != nil {
		stockTotal = 0
	}

	// Remaining equity goes to MF
	mfAmount := equityAmount - stockTotal

	// 6. MF allocation
	if mfAmount > 0 {
		mfRecs := a.distributeMFAllocation(mfAmount, niftyMetrics, regime)
		recommendations = append(recommendations, mfRecs...)
	}

	// 7. Stock allocations
	if len(stockRecs) > 0 {
		recommendations = append(recommendations, stockRecs...)
	}

	// 8. Debt allocation
	if debtAmount > 0 {
		recommendations = append(recommendations, models.AllocationRecommendation{
			AssetSymbol: "DEBT_BUCKET",
			AssetType:   models.AssetTypeDebt,
			Amount:      debtAmount,
			Source:      models.AllocationSourceSIP,
			Reason:      reasonDebtSIP(allocation.Debt, regime),
		})
	}

	return recommendations, nil
}

// PEFearBoost converts Nifty PE into an additive adjustment to DMA distance.
// Low PE increases effective fear (positive boost); high PE reduces it.
func PEFearBoost(niftyPE float64) float64 {
	if niftyPE <= 0 {
		return 0
	}
	return (PENeutralReference - niftyPE) * PEFearBoostWeight
}

// EffectiveDMADistance combines 200-DMA distance with a PE-based fear boost for regime detection.
func EffectiveDMADistance(dmaDistance, niftyPE float64) float64 {
	return dmaDistance + PEFearBoost(niftyPE)
}

func marketRegimeFromDMADistance(effectiveDistance float64) MarketRegime {
	if effectiveDistance >= DeepFearThreshold {
		return RegimeDeepFear
	}
	if math.Abs(effectiveDistance) <= NeutralUpperBound {
		return RegimeNeutral
	}
	if effectiveDistance > 0 {
		return RegimeFear
	}
	return RegimeGreed
}

// DetermineMarketRegime determines the current market regime from Nifty 200-DMA distance
// and Nifty PE. PE is weighted heavily via EffectiveDMADistance. Pass niftyPE <= 0 to
// skip PE adjustment when PE is unavailable.
func (a *allocator) DetermineMarketRegime(niftyMetrics *models.DropMetrics, niftyPE float64) MarketRegime {
	return marketRegimeFromDMADistance(EffectiveDMADistance(niftyMetrics.DMADistance, niftyPE))
}

// CalculateDropMetrics calculates all drop metrics for a given symbol.
// All independent API calls are made concurrently for performance.
func (a *allocator) CalculateDropMetrics(symbol string) (*models.DropMetrics, error) {
	fetcher := a.fetcher

	var (
		price, dma200, weekAgoPrice, monthAgoPrice, high52w float64
		priceErr, dmaErr, weekErr, monthErr, highErr        error
		wg                                                  sync.WaitGroup
	)

	wg.Add(5)
	go func() { defer wg.Done(); price, priceErr = fetcher.FetchCurrentPrice(symbol) }()
	go func() { defer wg.Done(); dma200, dmaErr = fetcher.FetchDMA200(symbol) }()
	go func() { defer wg.Done(); weekAgoPrice, weekErr = fetcher.FetchPriceNDaysAgo(symbol, 7) }()
	go func() { defer wg.Done(); monthAgoPrice, monthErr = fetcher.FetchPriceNDaysAgo(symbol, 22) }()
	go func() { defer wg.Done(); high52w, highErr = fetcher.Fetch52WeekHigh(symbol) }()
	wg.Wait()

	if priceErr != nil {
		return nil, fmt.Errorf("failed to fetch current price for %s: %v", symbol, priceErr)
	}

	metrics := &models.DropMetrics{
		Symbol:       symbol,
		CurrentPrice: price,
	}

	if dmaErr != nil {
		dma200 = price
	}
	metrics.DMA200 = dma200
	if dma200 > 0 {
		metrics.DMADistance = (dma200 - price) / dma200 * 100
	}

	if weekErr != nil {
		weekAgoPrice = price
	}
	metrics.WeekAgoPrice = weekAgoPrice
	if weekAgoPrice > 0 {
		metrics.WeekDrop = (weekAgoPrice - price) / weekAgoPrice * 100
	}

	if monthErr != nil {
		monthAgoPrice = price
	}
	metrics.MonthAgoPrice = monthAgoPrice
	if monthAgoPrice > 0 {
		metrics.MonthDrop = (monthAgoPrice - price) / monthAgoPrice * 100
	}

	if highErr != nil {
		high52w = price
	}
	metrics.RecentHighPrice = high52w
	if high52w > 0 {
		metrics.HighDrop = (high52w - price) / high52w * 100
	}

	metrics.CalculateCompositeScore()
	return metrics, nil
}

// QualifiedStock represents a stock that qualifies for investment
type QualifiedStock struct {
	Symbol       string
	Metrics      *models.DropMetrics
	MarketCap    float64
	RiskWeight   float64 // 1.0 for large cap, 0.7 for mid, 0.4 for small
	QualifyScore float64 // How much "too down" the stock is
}

// stockAnalysisResult holds the result of analyzing a single stock concurrently
type stockAnalysisResult struct {
	Stock     QualifiedStock
	Qualifies bool
}

// distributeToStocks determines which stocks qualify and how much to allocate.
// All stock metrics and market caps are fetched concurrently.
func (a *allocator) distributeToStocks(equityAmount float64, niftyMetrics *models.DropMetrics, regime MarketRegime) ([]models.AllocationRecommendation, float64, error) {
	results := make([]stockAnalysisResult, len(StockWatchlist))
	var wg sync.WaitGroup

	for i, symbol := range StockWatchlist {
		wg.Add(1)
		go func(idx int, sym string) {
			defer wg.Done()

			metrics, err := a.CalculateDropMetrics(sym)
			if err != nil {
				return
			}

			qualifies, _, qualifyScore := a.stockQualifies(metrics, niftyMetrics)
			if !qualifies {
				return
			}

			marketCap, err := a.fetcher.FetchMarketCap(sym)
			if err != nil || marketCap < 10000000000 {
				if fallback, ok := FallbackMarketCaps[sym]; ok {
					marketCap = fallback
				}
			}

			results[idx] = stockAnalysisResult{
				Qualifies: true,
				Stock: QualifiedStock{
					Symbol:       sym,
					Metrics:      metrics,
					MarketCap:    marketCap,
					RiskWeight:   calculateRiskWeight(marketCap),
					QualifyScore: qualifyScore,
				},
			}
		}(i, symbol)
	}
	wg.Wait()

	var qualifiedStocks []QualifiedStock
	for _, r := range results {
		if r.Qualifies {
			qualifiedStocks = append(qualifiedStocks, r.Stock)
		}
	}

	if len(qualifiedStocks) == 0 {
		return nil, 0, nil
	}

	// Sort by qualify score (most down first), then by risk weight (safer first)
	sort.Slice(qualifiedStocks, func(i, j int) bool {
		// Primary: qualify score (higher = more down = better)
		// Secondary: risk weight (higher = less risky = better)
		if math.Abs(qualifiedStocks[i].QualifyScore-qualifiedStocks[j].QualifyScore) > 0.1 {
			return qualifiedStocks[i].QualifyScore > qualifiedStocks[j].QualifyScore
		}
		return qualifiedStocks[i].RiskWeight > qualifiedStocks[j].RiskWeight
	})

	// Calculate total allocation to stocks
	// In Fear/DeepFear: Less to stocks (prefer MF), unless stocks are exceptionally down
	// In Neutral/Greed: Stocks get minimum allocation
	stockAllocationPercent := a.calculateStockAllocationPercent(qualifiedStocks, regime)
	totalStockAmount := equityAmount * stockAllocationPercent

	// Distribute among qualified stocks and get actual allocated amount
	recs, actualAllocated := a.allocateToQualifiedStocks(qualifiedStocks, totalStockAmount, niftyMetrics)
	return recs, actualAllocated, nil
}

// stockQualifies checks if a stock is "too down" to qualify for investment
func (a *allocator) stockQualifies(stockMetrics, niftyMetrics *models.DropMetrics) (bool, string, float64) {
	qualifyScore := 0.0
	reasons := []string{}

	// Criterion 1: Stock drop is >= 2x Nifty drop (from 200 DMA)
	if niftyMetrics.DMADistance > 0 && stockMetrics.DMADistance >= niftyMetrics.DMADistance*StockDropMultiplier {
		qualifyScore += stockMetrics.DMADistance / niftyMetrics.DMADistance
		reasons = append(reasons, fmt.Sprintf("DMA drop %.1fx Nifty", stockMetrics.DMADistance/niftyMetrics.DMADistance))
	}

	// Criterion 2: Sharp weekly drop (>10%)
	if stockMetrics.WeekDrop >= SharpWeeklyDrop {
		qualifyScore += stockMetrics.WeekDrop / 10.0 // Normalize
		reasons = append(reasons, fmt.Sprintf("Sharp week drop %.1f%%", stockMetrics.WeekDrop))
	}

	// Criterion 3: Sharp monthly drop (>15%)
	if stockMetrics.MonthDrop >= SharpMonthlyDrop {
		qualifyScore += stockMetrics.MonthDrop / 15.0 // Normalize
		reasons = append(reasons, fmt.Sprintf("Sharp month drop %.1f%%", stockMetrics.MonthDrop))
	}

	// Criterion 4: Stock is significantly more down than Nifty overall (2x composite score)
	if niftyMetrics.CompositeScore > 0 && stockMetrics.CompositeScore >= niftyMetrics.CompositeScore*StockDropMultiplier {
		qualifyScore += stockMetrics.CompositeScore / niftyMetrics.CompositeScore
		reasons = append(reasons, fmt.Sprintf("Composite %.1fx Nifty", stockMetrics.CompositeScore/niftyMetrics.CompositeScore))
	}

	// Stock qualifies if it meets at least one criterion
	qualifies := len(reasons) > 0

	var reason strings.Builder
	if qualifies {
		for i, r := range reasons {
			if i > 0 {
				reason.WriteString(", ")
			}
			reason.WriteString(r)
		}
	}

	return qualifies, reason.String(), qualifyScore
}

// calculateRiskWeight returns a weight based on market cap
// Lower weight = higher volatility/risk = less allocation
func calculateRiskWeight(marketCap float64) float64 {
	// If market cap is 0 or very low, it's likely a fetch failure
	// Use unknown weight (conservative)
	if marketCap < 10000000000 { // Less than 1000 Crore is suspicious for these stocks
		return UnknownCapWeight
	}
	if marketCap >= LargeCapThreshold {
		return LargeCapWeight
	} else if marketCap >= MidCapThreshold {
		return MidCapWeight
	}
	return SmallCapWeight
}

// calculateStockAllocationPercent determines what % of equity should go to stocks
func (a *allocator) calculateStockAllocationPercent(qualifiedStocks []QualifiedStock, regime MarketRegime) float64 {
	if len(qualifiedStocks) == 0 {
		return 0
	}

	// Base allocation depends on regime
	// In Fear markets: Prefer MF, so minimal to stocks
	// In Greed markets: Very minimal to stocks
	var basePercent float64
	switch regime {
	case RegimeDeepFear:
		basePercent = MinStockAllocation // Only 5% to stocks (prefer MF)
	case RegimeFear:
		basePercent = MinStockAllocation + 0.03 // 8% to stocks
	case RegimeNeutral:
		basePercent = MinStockAllocation + 0.07 // 12% to stocks
	case RegimeGreed:
		basePercent = MinStockAllocation // 5% to stocks
	default:
		basePercent = MinStockAllocation
	}

	// Boost allocation only if stocks are extremely down (4x+ more than Nifty)
	avgQualifyScore := 0.0
	for _, s := range qualifiedStocks {
		avgQualifyScore += s.QualifyScore
	}
	avgQualifyScore /= float64(len(qualifiedStocks))

	// If average qualify score is > 4 (meaning stocks are 4x+ more down), small boost
	if avgQualifyScore > 4.0 {
		boost := math.Min((avgQualifyScore-4.0)*0.02, 0.15) // Max 10% boost (conservative)
		basePercent += boost
	}

	// Cap at max
	if basePercent > MaxStockAllocation {
		basePercent = MaxStockAllocation
	}

	return basePercent
}

// allocateToQualifiedStocks distributes the stock amount among qualified stocks
// Returns recommendations and the actual total allocated (may be less than totalAmount due to caps)
func (a *allocator) allocateToQualifiedStocks(stocks []QualifiedStock, totalAmount float64, niftyMetrics *models.DropMetrics) ([]models.AllocationRecommendation, float64) {
	if len(stocks) == 0 || totalAmount <= 0 {
		return nil, 0
	}

	// Take top N stocks (max 5)
	maxStocks := min(len(stocks), 5)
	stocks = stocks[:maxStocks]

	// Calculate weighted scores for distribution
	totalWeightedScore := 0.0
	for _, s := range stocks {
		// Weight = QualifyScore * RiskWeight
		totalWeightedScore += s.QualifyScore * s.RiskWeight
	}

	var recs []models.AllocationRecommendation
	actualTotal := 0.0
	smallCapTotal := 0.0
	maxSmallCapAmount := totalAmount * MaxSmallCapTotalPercent

	for _, stock := range stocks {
		// Proportional allocation based on weighted score
		proportion := 0.0
		if totalWeightedScore > 0 {
			proportion = (stock.QualifyScore * stock.RiskWeight) / totalWeightedScore
		} else {
			proportion = 1.0 / float64(len(stocks))
		}

		allocation := totalAmount * proportion

		// Cap single stock allocation
		maxSingleAmount := totalAmount * MaxSingleStockPercent
		if allocation > maxSingleAmount {
			allocation = maxSingleAmount
		}

		// Determine cap category
		capLabel := "Large"
		isSmallCap := false
		switch stock.RiskWeight {
		case MidCapWeight:
			capLabel = "Mid"
		case SmallCapWeight:
			capLabel = "Small"
			isSmallCap = true
		case UnknownCapWeight:
			capLabel = "Unknown"
			isSmallCap = true // Treat unknown as small cap for capping purposes
		}

		// Apply small cap total allocation cap
		if isSmallCap {
			if smallCapTotal+allocation > maxSmallCapAmount {
				// Reduce allocation to fit within cap
				allocation = maxSmallCapAmount - smallCapTotal
				if allocation <= 0 {
					continue // Skip this stock, small cap quota exhausted
				}
			}
			smallCapTotal += allocation
		}

		actualTotal += allocation

		recs = append(recs, models.AllocationRecommendation{
			AssetSymbol: stock.Symbol,
			AssetType:   models.AssetTypeStock,
			Amount:      allocation,
			Source:      models.AllocationSourceSIP,
			Reason: reasonStock(
				capLabel,
				stock.MarketCap/10000000,
				stock.Metrics.DMADistance, niftyMetrics.DMADistance,
				stock.Metrics.WeekDrop, stock.Metrics.MonthDrop,
				stock.QualifyScore,
			),
		})
	}

	return recs, actualTotal
}

// distributeMFAllocation distributes MF allocation.
// MF metrics are fetched concurrently.
func (a *allocator) distributeMFAllocation(amount float64, niftyMetrics *models.DropMetrics, regime MarketRegime) []models.AllocationRecommendation {
	if len(MFWatchlist) == 0 || amount <= 0 {
		return nil
	}

	perMF := amount / float64(len(MFWatchlist))

	type mfResult struct {
		Code    string
		Metrics *models.DropMetrics
	}
	mfResults := make([]mfResult, len(MFWatchlist))
	var wg sync.WaitGroup
	for i, mfCode := range MFWatchlist {
		wg.Add(1)
		go func(idx int, code string) {
			defer wg.Done()
			metrics, _ := a.CalculateDropMetrics(code)
			mfResults[idx] = mfResult{Code: code, Metrics: metrics}
		}(i, mfCode)
	}
	wg.Wait()

	recs := make([]models.AllocationRecommendation, 0, len(MFWatchlist))
	for _, mr := range mfResults {
		var mfDMA *float64
		if mr.Metrics != nil {
			d := mr.Metrics.DMADistance
			mfDMA = &d
		}
		recs = append(recs, models.AllocationRecommendation{
			AssetSymbol: mr.Code,
			AssetType:   models.AssetTypeMF,
			Amount:      perMF,
			Source:      models.AllocationSourceSIP,
			Reason:      reasonMFSIP(regime, niftyMetrics.DMADistance, mfDMA),
		})
	}
	return recs
}

// isMF checks if a symbol is likely a Mutual Fund (numeric AMFI code)
func isMF(symbol string) bool {
	if symbol == "" {
		return false
	}
	for _, r := range symbol {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
