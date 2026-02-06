package allocator

import (
	"smart-alert/internal/domain/models"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDataFetcher
type MockDataFetcher struct {
	mock.Mock
}

func (m *MockDataFetcher) FetchCurrentPrice(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDataFetcher) FetchHistoricalData(symbol string, rangeStr string, interval string) ([]int64, []float64, error) {
	args := m.Called(symbol, rangeStr, interval)
	return args.Get(0).([]int64), args.Get(1).([]float64), args.Error(2)
}

func (m *MockDataFetcher) FetchDMA200(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDataFetcher) FetchMarketCap(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDataFetcher) FetchPE(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDataFetcher) FetchPriceNDaysAgo(symbol string, days int) (float64, error) {
	args := m.Called(symbol, days)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDataFetcher) Fetch52WeekHigh(symbol string) (float64, error) {
	args := m.Called(symbol)
	return args.Get(0).(float64), args.Error(1)
}

// Helper to setup Nifty mocks
func setupNiftyMocks(m *MockDataFetcher, currentPrice, dma200, weekAgo, monthAgo, high float64, pe float64) {
	m.On("FetchCurrentPrice", "^NSEI").Return(currentPrice, nil)
	m.On("FetchDMA200", "^NSEI").Return(dma200, nil)
	m.On("FetchPriceNDaysAgo", "^NSEI", 7).Return(weekAgo, nil)
	m.On("FetchPriceNDaysAgo", "^NSEI", 22).Return(monthAgo, nil)
	m.On("Fetch52WeekHigh", "^NSEI").Return(high, nil)
	m.On("FetchPE", "^NSEI").Return(pe, nil)
}

// Helper to setup stock mock that doesn't qualify (not down enough)
func setupNonQualifyingStockMocks(m *MockDataFetcher, symbol string) {
	m.On("FetchCurrentPrice", symbol).Return(100.0, nil)
	m.On("FetchDMA200", symbol).Return(100.0, nil) // At DMA
	m.On("FetchPriceNDaysAgo", symbol, 7).Return(100.0, nil)
	m.On("FetchPriceNDaysAgo", symbol, 22).Return(100.0, nil)
	m.On("Fetch52WeekHigh", symbol).Return(100.0, nil)
	m.On("FetchMarketCap", symbol).Return(500000000000.0, nil) // Mid cap
}

// Helper to setup stock mock that qualifies (very down)
func setupQualifyingStockMocks(m *MockDataFetcher, symbol string, marketCap float64) {
	// Stock is 25% below DMA (vs Nifty 10%) - qualifies as 2.5x
	m.On("FetchCurrentPrice", symbol).Return(75.0, nil)
	m.On("FetchDMA200", symbol).Return(100.0, nil)           // 25% below DMA
	m.On("FetchPriceNDaysAgo", symbol, 7).Return(85.0, nil)  // 11.7% weekly drop
	m.On("FetchPriceNDaysAgo", symbol, 22).Return(95.0, nil) // 21% monthly drop
	m.On("Fetch52WeekHigh", symbol).Return(120.0, nil)       // 37.5% from high
	m.On("FetchMarketCap", symbol).Return(marketCap, nil)
}

// Helper to setup MF mocks
func setupMFMocks(m *MockDataFetcher, symbol string) {
	m.On("FetchCurrentPrice", symbol).Return(50.0, nil)
	m.On("FetchDMA200", symbol).Return(55.0, nil)
	m.On("FetchPriceNDaysAgo", symbol, 7).Return(52.0, nil)
	m.On("FetchPriceNDaysAgo", symbol, 22).Return(54.0, nil)
	m.On("Fetch52WeekHigh", symbol).Return(60.0, nil)
}

func TestDetermineMarketRegime(t *testing.T) {
	tests := []struct {
		name           string
		dmaDistance    float64 // positive = below DMA
		expectedRegime MarketRegime
	}{
		{"Deep Fear - 15% below DMA", 15.0, RegimeDeepFear},
		{"Deep Fear - 10% below DMA", 10.0, RegimeDeepFear},
		{"Fear - 8% below DMA", 8.0, RegimeFear},
		{"Fear - 6% below DMA", 6.0, RegimeFear},
		{"Neutral - 5% below DMA", 5.0, RegimeNeutral},
		{"Neutral - 3% below DMA", 3.0, RegimeNeutral},
		{"Neutral - 1% below DMA", 1.0, RegimeNeutral},
		{"Neutral - at DMA", 0.0, RegimeNeutral},
		{"Neutral - 3% above DMA", -3.0, RegimeNeutral},
		{"Neutral - 5% above DMA", -5.0, RegimeNeutral},
		{"Greed - 6% above DMA", -6.0, RegimeGreed},
		{"Greed - 10% above DMA", -10.0, RegimeGreed},
	}

	mockFetcher := new(MockDataFetcher)
	alloc := NewAllocator(mockFetcher)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &models.DropMetrics{
				DMADistance: tt.dmaDistance,
			}
			regime := alloc.DetermineMarketRegime(metrics)
			assert.Equal(t, tt.expectedRegime, regime)
		})
	}
}

func TestAllocate_DeepFear_WithQualifyingStock(t *testing.T) {
	mockFetcher := new(MockDataFetcher)
	allocator := NewAllocator(mockFetcher)

	// Setup Nifty - 15% below DMA (Clearly Deep Fear: > 10% threshold)
	// DMADistance = (16000 - 13600) / 16000 * 100 = 15%
	setupNiftyMocks(mockFetcher, 13600, 16000, 14500, 15000, 17000, 22.0)

	// Setup one qualifying stock (RELIANCE - Large Cap, very down)
	setupQualifyingStockMocks(mockFetcher, "RELIANCE.NS", 1500000000000) // Large cap

	// Setup non-qualifying stocks
	for _, stock := range []string{"CIPLA.NS", "HAVELLS.NS", "ICICIBANK.NS", "SHRIRAMFIN.NS", "RAINBOW.NS", "KPITTECH.NS", "SENCO.NS", "ZAGGLE.NS", "CAMS.NS", "BAJFINANCE.NS"} {
		setupNonQualifyingStockMocks(mockFetcher, stock)
	}

	// Setup MF
	setupMFMocks(mockFetcher, "150346")

	recs, err := allocator.Allocate(10000, nil, 0, nil)

	assert.NoError(t, err)

	// In Deep Fear: 100% Equity, 0% Debt
	// Within Equity: Most to MF, some to qualifying stock
	var mfTotal, stockTotal, debtTotal float64
	for _, r := range recs {
		switch r.AssetType {
		case models.AssetTypeMF:
			mfTotal += r.Amount
		case models.AssetTypeStock:
			stockTotal += r.Amount
		case models.AssetTypeDebt:
			debtTotal += r.Amount
		}
	}

	// Debt should be 0% = 0
	assert.Equal(t, 0.0, debtTotal, "Debt should be 0% in Deep Fear")

	// Total equity should be 100% = 10000
	totalEquity := mfTotal + stockTotal
	assert.InDelta(t, 10000.0, totalEquity, 1.0, "Total equity should be 100%")

	// MF should be majority of equity (stocks get only ~10-15% of equity in Deep Fear)
	assert.Greater(t, mfTotal, stockTotal, "MF should be greater than stocks in Deep Fear")
}

func TestAllocate_Greed_AllToDebt(t *testing.T) {
	mockFetcher := new(MockDataFetcher)
	allocator := NewAllocator(mockFetcher)

	// Setup Nifty - 10% above DMA (Greed)
	setupNiftyMocks(mockFetcher, 17600, 16000, 17400, 17200, 18000, 25.0)

	// Setup all stocks as non-qualifying (not down enough)
	for _, stock := range StockWatchlist {
		setupNonQualifyingStockMocks(mockFetcher, stock)
	}

	// Setup MF
	setupMFMocks(mockFetcher, "150346")

	recs, err := allocator.Allocate(10000, nil, 0, nil)

	assert.NoError(t, err)

	// In Greed: 40% Equity, 60% Debt
	var mfTotal, stockTotal, debtTotal float64
	for _, r := range recs {
		switch r.AssetType {
		case models.AssetTypeMF:
			mfTotal += r.Amount
		case models.AssetTypeStock:
			stockTotal += r.Amount
		case models.AssetTypeDebt:
			debtTotal += r.Amount
		}
	}

	// Debt should be ~60% = 6000
	assert.InDelta(t, 6000.0, debtTotal, 100.0, "Debt should be ~60% in Greed regime")

	// Total equity should be ~40% = 4000
	totalEquity := mfTotal + stockTotal
	assert.InDelta(t, 4000.0, totalEquity, 100.0, "Total equity should be ~40%")

	// No qualifying stocks, so all equity goes to MF
	assert.Equal(t, 0.0, stockTotal, "No stocks should qualify in Greed with non-down stocks")
	assert.InDelta(t, 4000.0, mfTotal, 100.0, "All equity should go to MF")
}

func TestAllocate_Fear_StockQualifiesSharpDrop(t *testing.T) {
	mockFetcher := new(MockDataFetcher)
	allocator := NewAllocator(mockFetcher)

	// Setup Nifty - 8% below DMA (Fear: >5% but <10%)
	// DMADistance = (16000 - 14720) / 16000 * 100 = 8%
	setupNiftyMocks(mockFetcher, 14720, 16000, 15200, 15500, 17000, 20.0)

	// Setup RELIANCE with sharp weekly drop (qualifies)
	mockFetcher.On("FetchCurrentPrice", "RELIANCE.NS").Return(85.0, nil)
	mockFetcher.On("FetchDMA200", "RELIANCE.NS").Return(100.0, nil)
	mockFetcher.On("FetchPriceNDaysAgo", "RELIANCE.NS", 7).Return(100.0, nil) // 15% week drop (sharp!)
	mockFetcher.On("FetchPriceNDaysAgo", "RELIANCE.NS", 22).Return(105.0, nil)
	mockFetcher.On("Fetch52WeekHigh", "RELIANCE.NS").Return(110.0, nil)
	mockFetcher.On("FetchMarketCap", "RELIANCE.NS").Return(1500000000000.0, nil)

	// Setup other stocks as non-qualifying
	for _, stock := range []string{"CIPLA.NS", "HAVELLS.NS", "ICICIBANK.NS", "SHRIRAMFIN.NS", "RAINBOW.NS", "KPITTECH.NS", "SENCO.NS", "ZAGGLE.NS", "CAMS.NS", "BAJFINANCE.NS"} {
		setupNonQualifyingStockMocks(mockFetcher, stock)
	}

	// Setup MF
	setupMFMocks(mockFetcher, "150346")

	recs, err := allocator.Allocate(10000, nil, 0, nil)

	assert.NoError(t, err)

	// Should have stock allocation (RELIANCE qualifies due to sharp weekly drop)
	foundReliance := false
	for _, r := range recs {
		if r.AssetSymbol == "RELIANCE.NS" {
			foundReliance = true
			assert.Greater(t, r.Amount, 0.0, "RELIANCE should have allocation")
		}
	}
	assert.True(t, foundReliance, "RELIANCE should qualify due to sharp weekly drop")
}

func TestCalculateDropMetrics(t *testing.T) {
	mockFetcher := new(MockDataFetcher)
	allocator := NewAllocator(mockFetcher)

	setupNiftyMocks(mockFetcher, 15000, 16000, 15500, 16000, 17000, 21.0)

	metrics, err := allocator.CalculateDropMetrics("^NSEI")

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, "^NSEI", metrics.Symbol)
	assert.Equal(t, 15000.0, metrics.CurrentPrice)
	assert.Equal(t, 16000.0, metrics.DMA200)

	// DMA Distance: (16000 - 15000) / 16000 * 100 = 6.25%
	assert.InDelta(t, 6.25, metrics.DMADistance, 0.1)

	// Week Drop: (15500 - 15000) / 15500 * 100 = 3.22%
	assert.InDelta(t, 3.22, metrics.WeekDrop, 0.1)

	// Month Drop: (16000 - 15000) / 16000 * 100 = 6.25%
	assert.InDelta(t, 6.25, metrics.MonthDrop, 0.1)

	// High Drop: (17000 - 15000) / 17000 * 100 = 11.76%
	assert.InDelta(t, 11.76, metrics.HighDrop, 0.1)
}

func TestRiskWeightByMarketCap(t *testing.T) {
	tests := []struct {
		name       string
		marketCap  float64
		wantWeight float64
	}{
		{"Large Cap", 1500000000000, LargeCapWeight},
		{"Large Cap Threshold", 1000000000000, LargeCapWeight},
		{"Mid Cap", 500000000000, MidCapWeight},
		{"Mid Cap Threshold", 300000000000, MidCapWeight},
		{"Small Cap", 100000000000, SmallCapWeight},
		{"Small Cap Very Small", 50000000000, SmallCapWeight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := calculateRiskWeight(tt.marketCap)
			assert.Equal(t, tt.wantWeight, weight)
		})
	}
}

func TestAllocate_DeepFear_DeploysDebtReserve(t *testing.T) {
	mockFetcher := new(MockDataFetcher)
	allocator := NewAllocator(mockFetcher)

	// Setup Nifty - 15% below DMA (Deep Fear)
	setupNiftyMocks(mockFetcher, 13600, 16000, 14500, 15000, 17000, 22.0)

	// Setup one qualifying stock
	setupQualifyingStockMocks(mockFetcher, "RELIANCE.NS", 1500000000000)

	// Setup non-qualifying stocks
	for _, stock := range []string{"CIPLA.NS", "HAVELLS.NS", "ICICIBANK.NS", "SHRIRAMFIN.NS", "RAINBOW.NS", "KPITTECH.NS", "SENCO.NS", "ZAGGLE.NS", "CAMS.NS", "BAJFINANCE.NS"} {
		setupNonQualifyingStockMocks(mockFetcher, stock)
	}

	// Setup MF
	setupMFMocks(mockFetcher, "150346")

	// Allocate with debt reserve
	debtReserve := 100000.0 // ₹1 lakh debt reserve
	recs, err := allocator.Allocate(10000, nil, debtReserve, nil)

	assert.NoError(t, err)

	// Find panic buy recommendations
	var niftyETFAmount, whiteoakAmount float64
	for _, r := range recs {
		if r.AssetSymbol == Nifty50ETFSymbol {
			niftyETFAmount = r.Amount
		}
		if r.AssetSymbol == WhiteoakFlexiCapCode && strings.HasPrefix(r.Reason, "PANIC BUY") {
			whiteoakAmount = r.Amount
		}
	}

	// 50% of debt reserve = 50000, split 50-50 = 25000 each
	expectedDeployment := debtReserve * DebtReserveDeploymentPct / 2
	assert.InDelta(t, expectedDeployment, niftyETFAmount, 1.0, "Nifty50 ETF should get 50% of deployment")
	assert.InDelta(t, expectedDeployment, whiteoakAmount, 1.0, "Whiteoak should get 50% of deployment")
}

func TestAllocate_PETrigger_Deploys30Percent(t *testing.T) {
	mockFetcher := new(MockDataFetcher)
	allocator := NewAllocator(mockFetcher)

	// Setup Nifty - Neutral regime (no panic buy from DMA), but PE = 18.0 (Trigger!)
	setupNiftyMocks(mockFetcher, 16000, 16000, 16000, 16000, 16500, 18.0)

	// Setup non-qualifying stocks
	for _, stock := range StockWatchlist {
		setupNonQualifyingStockMocks(mockFetcher, stock)
	}

	// Setup MF
	setupMFMocks(mockFetcher, "150346")

	// Allocate with debt reserve
	debtReserve := 100000.0 // ₹1 lakh
	recs, err := allocator.Allocate(10000, nil, debtReserve, nil)

	assert.NoError(t, err)

	// Verify PE trigger deployment
	var peTriggerAmount float64
	foundPETrigger := false
	for _, r := range recs {
		if r.AssetSymbol == Nifty50ETFSymbol && strings.HasPrefix(r.Reason, "PE_TRIGGER") {
			peTriggerAmount = r.Amount
			foundPETrigger = true
		}
	}

	assert.True(t, foundPETrigger, "Should have PE trigger recommendation")
	// 30% of 100000 = 30000
	assert.InDelta(t, 30000.0, peTriggerAmount, 1.0, "PE trigger should deploy 30% of debt reserve")
}
