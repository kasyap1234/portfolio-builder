package allocator

import "fmt"

func reasonPanicBuyETF(deployPct, totalDeploy, amount float64, regimeLabel string, niftyDMA float64) string {
	return fmt.Sprintf(
		"Panic deploy: %.0f%% of debt reserve (%s total). %s — Nifty %.1f%% below 200-DMA. This leg: Nifty 50 ETF (%s).",
		deployPct*100, formatINR(totalDeploy), regimeLabel, niftyDMA, formatINR(amount),
	)
}

func reasonPanicBuyMF(deployPct, totalDeploy, amount float64, regimeLabel string, niftyDMA float64) string {
	return fmt.Sprintf(
		"Panic deploy: %.0f%% of debt reserve (%s total). %s — Nifty %.1f%% below 200-DMA. This leg: Whiteoak Flexi Cap (%s).",
		deployPct*100, formatINR(totalDeploy), regimeLabel, niftyDMA, formatINR(amount),
	)
}

func reasonPETrigger(niftyPE, threshold, deployPct, amount float64) string {
	return fmt.Sprintf(
		"Nifty PE %.1f is at or below %.1f. Deploy %.0f%% of debt reserve into Nifty 50 ETF (%s).",
		niftyPE, threshold, deployPct*100, formatINR(amount),
	)
}

func reasonDebtSIP(debtPct float64, regime MarketRegime) string {
	return fmt.Sprintf(
		"%.0f%% of monthly SIP to debt. %s regime allocation is %.0f%% equity / %.0f%% debt.",
		debtPct*100, regime, (1-debtPct)*100, debtPct*100,
	)
}

func reasonMFSIP(regime MarketRegime, niftyDMA float64, mfDMA *float64) string {
	s := fmt.Sprintf("Monthly SIP equity. %s regime — Nifty %.1f%% below 200-DMA.", regime, niftyDMA)
	if mfDMA != nil {
		s += fmt.Sprintf(" Fund %.1f%% below 200-DMA.", *mfDMA)
	}
	return s
}

func reasonStock(capLabel string, marketCapCr, stockDMA, niftyDMA, weekDrop, monthDrop, score float64) string {
	return fmt.Sprintf(
		"%s cap (~₹%.0f Cr). Beaten Nifty on drawdown: DMA %.1f%% vs %.1f%%, week %.1f%%, month %.1f%%. Score %.1f.",
		capLabel, marketCapCr, stockDMA, niftyDMA, weekDrop, monthDrop, score,
	)
}

func formatINR(amount float64) string {
	return fmt.Sprintf("₹%.0f", amount)
}
