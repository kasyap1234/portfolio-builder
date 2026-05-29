package allocationreport

import (
	"strings"
	"testing"

	"smart-alert/internal/domain/models"
)

func TestFormatConsoleSIPAndReserve(t *testing.T) {
	recs := []models.AllocationRecommendation{
		{AssetSymbol: "150346", AssetType: models.AssetTypeMF, Amount: 36000, Source: models.AllocationSourceSIP, Reason: "Monthly SIP equity."},
		{AssetSymbol: "DEBT_BUCKET", AssetType: models.AssetTypeDebt, Amount: 9000, Source: models.AllocationSourceSIP, Reason: "20% to debt."},
		{AssetSymbol: "NIFTY50-ETF", AssetType: models.AssetTypeETF, Amount: 1200, Source: models.AllocationSourceDebtReserve, Reason: "Panic deploy."},
	}
	out := FormatConsole(recs, Options{SIPAmount: 45000, DebtReserve: 120000, Regime: "FEAR"})
	if !strings.Contains(out, "Monthly SIP") {
		t.Fatalf("expected SIP section: %s", out)
	}
	if !strings.Contains(out, "Debt reserve") {
		t.Fatalf("expected reserve section: %s", out)
	}
	if !strings.Contains(out, "100%") {
		t.Fatalf("expected SIP summary: %s", out)
	}
}
