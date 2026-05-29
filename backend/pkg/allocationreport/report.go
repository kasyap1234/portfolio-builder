package allocationreport

import (
	"fmt"
	"strings"

	"smart-alert/internal/domain/models"
)

// Options configures allocation report output.
type Options struct {
	SIPAmount   float64
	DebtReserve float64 // optional; 0 hides reserve context in header
	Regime      string
	NiftyPE     float64
	Title       string // optional; default "Monthly SIP Allocation"
}

// assetLabels maps internal symbols to readable names.
var assetLabels = map[string]string{
	"150346":      "Whiteoak Flexi Cap",
	"NIFTY50-ETF": "Nifty 50 BeES (ETF)",
	"DEBT_BUCKET": "Debt / liquid funds",
}

// AssetLabel returns a human-readable name for a symbol.
func AssetLabel(symbol string) string {
	if label, ok := assetLabels[symbol]; ok {
		return label
	}
	return symbol
}

type bucket struct {
	title string
	recs  []models.AllocationRecommendation
	total float64
}

type totals struct {
	sipMF, sipETF, sipStock, sipDebt float64
	reserveTotal                     float64
	grandTotal                       float64
}

func classify(recs []models.AllocationRecommendation) (sip []bucket, reserve []bucket, t totals) {
	sipByType := map[models.AssetType]*bucket{
		models.AssetTypeMF:    {title: "Mutual funds"},
		models.AssetTypeETF:   {title: "ETFs"},
		models.AssetTypeStock: {title: "Individual stocks"},
		models.AssetTypeDebt:  {title: "Debt"},
	}
	reserveByType := map[models.AssetType]*bucket{
		models.AssetTypeMF:    {title: "Mutual funds"},
		models.AssetTypeETF:   {title: "ETFs"},
		models.AssetTypeStock: {title: "Individual stocks"},
	}

	for _, r := range recs {
		t.grandTotal += r.Amount
		fromReserve := r.Source == models.AllocationSourceDebtReserve

		if fromReserve {
			t.reserveTotal += r.Amount
			b := reserveByType[r.AssetType]
			if b != nil {
				b.recs = append(b.recs, r)
				b.total += r.Amount
			}
			continue
		}

		switch r.AssetType {
		case models.AssetTypeMF:
			t.sipMF += r.Amount
		case models.AssetTypeETF:
			t.sipETF += r.Amount
		case models.AssetTypeStock:
			t.sipStock += r.Amount
		case models.AssetTypeDebt:
			t.sipDebt += r.Amount
		}
		b := sipByType[r.AssetType]
		if b != nil {
			b.recs = append(b.recs, r)
			b.total += r.Amount
		}
	}

	order := []models.AssetType{models.AssetTypeMF, models.AssetTypeETF, models.AssetTypeStock, models.AssetTypeDebt}
	for _, typ := range order {
		if b := sipByType[typ]; b != nil && len(b.recs) > 0 {
			sip = append(sip, *b)
		}
		if b := reserveByType[typ]; b != nil && len(b.recs) > 0 {
			reserve = append(reserve, *b)
		}
	}
	return sip, reserve, t
}

func pct(amount, base float64) float64 {
	if base <= 0 {
		return 0
	}
	return amount / base * 100
}

func formatINR(amount float64) string {
	return fmt.Sprintf("₹%s", commaAmount(amount))
}

func commaAmount(amount float64) string {
	n := int64(amount + 0.5)
	if n < 0 {
		return fmt.Sprintf("%d", n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, ",")
}

func formatLine(symbol string, amount float64, reason string, indent string) string {
	label := AssetLabel(symbol)
	var b strings.Builder
	if label != symbol {
		b.WriteString(fmt.Sprintf("%s%s · %s\n", indent, symbol, label))
	} else {
		b.WriteString(fmt.Sprintf("%s%s\n", indent, symbol))
	}
	b.WriteString(fmt.Sprintf("%s  %s\n", indent, formatINR(amount)))
	if reason != "" {
		for _, line := range wrapText(reason, 68) {
			b.WriteString(fmt.Sprintf("%s  %s\n", indent, line))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len() == 0 {
			cur.WriteString(w)
			continue
		}
		if cur.Len()+1+len(w) > width {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
		} else {
			cur.WriteString(" ")
			cur.WriteString(w)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

// FormatConsole renders a plain-text report for terminals and logs.
func FormatConsole(recs []models.AllocationRecommendation, opts Options) string {
	title := opts.Title
	if title == "" {
		title = "Monthly SIP Allocation"
	}
	sipBuckets, reserveBuckets, t := classify(recs)
	sipEquity := t.sipMF + t.sipETF + t.sipStock

	var sb strings.Builder
	sb.WriteString(strings.Repeat("═", 44) + "\n")
	sb.WriteString(fmt.Sprintf(" %s\n", title))
	sb.WriteString(strings.Repeat("═", 44) + "\n")
	sb.WriteString(fmt.Sprintf(" SIP budget:  %s\n", formatINR(opts.SIPAmount)))
	if opts.DebtReserve > 0 {
		sb.WriteString(fmt.Sprintf(" Debt reserve: %s (panic buys are extra)\n", formatINR(opts.DebtReserve)))
	}
	sb.WriteString(fmt.Sprintf(" Regime:       %s\n", opts.Regime))
	if opts.NiftyPE > 0 {
		sb.WriteString(fmt.Sprintf(" Nifty PE:     %.2f\n", opts.NiftyPE))
	}

	sb.WriteString("\n")
	sb.WriteString("▸ Monthly SIP\n")
	sb.WriteString(strings.Repeat("─", 44) + "\n")
	if len(sipBuckets) == 0 {
		sb.WriteString("  (no SIP allocations)\n")
	} else {
		for _, b := range sipBuckets {
			sb.WriteString(fmt.Sprintf("\n  %s — %s\n", b.title, formatINR(b.total)))
			for i, r := range b.recs {
				if i > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(formatLine(r.AssetSymbol, r.Amount, r.Reason, "  "))
				sb.WriteString("\n")
			}
		}
	}

	if len(reserveBuckets) > 0 {
		sb.WriteString("\n")
		sb.WriteString("▸ Debt reserve (additional, not from SIP)\n")
		sb.WriteString(strings.Repeat("─", 44) + "\n")
		for _, b := range reserveBuckets {
			sb.WriteString(fmt.Sprintf("\n  %s — %s\n", b.title, formatINR(b.total)))
			for i, r := range b.recs {
				if i > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(formatLine(r.AssetSymbol, r.Amount, r.Reason, "  "))
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString("\n")
	sb.WriteString("▸ Summary\n")
	sb.WriteString(strings.Repeat("─", 44) + "\n")
	sb.WriteString(fmt.Sprintf(" SIP deployed:     %s  (100%%)\n", formatINR(opts.SIPAmount)))
	sb.WriteString(fmt.Sprintf("   Equity:         %s  (%4.1f%%)\n", formatINR(sipEquity), pct(sipEquity, opts.SIPAmount)))
	if t.sipMF > 0 {
		sb.WriteString(fmt.Sprintf("     Mutual funds: %s  (%4.1f%%)\n", formatINR(t.sipMF), pct(t.sipMF, opts.SIPAmount)))
	}
	if t.sipETF > 0 {
		sb.WriteString(fmt.Sprintf("     ETFs:         %s  (%4.1f%%)\n", formatINR(t.sipETF), pct(t.sipETF, opts.SIPAmount)))
	}
	if t.sipStock > 0 {
		sb.WriteString(fmt.Sprintf("     Stocks:       %s  (%4.1f%%)\n", formatINR(t.sipStock), pct(t.sipStock, opts.SIPAmount)))
	}
	sb.WriteString(fmt.Sprintf("   Debt:           %s  (%4.1f%%)\n", formatINR(t.sipDebt), pct(t.sipDebt, opts.SIPAmount)))
	if t.reserveTotal > 0 {
		sb.WriteString(fmt.Sprintf("\n Reserve deployed: %s  (on top of SIP)\n", formatINR(t.reserveTotal)))
		sb.WriteString(fmt.Sprintf(" Cash needed:      %s\n", formatINR(t.grandTotal)))
	}

	return sb.String()
}

// FormatTelegram renders a Markdown message for Telegram.
func FormatTelegram(recs []models.AllocationRecommendation, opts Options) string {
	sipBuckets, reserveBuckets, t := classify(recs)
	sipEquity := t.sipMF + t.sipETF + t.sipStock

	var sb strings.Builder
	sb.WriteString("*Monthly SIP Allocation*\n\n")
	sb.WriteString(fmt.Sprintf("SIP: *%s* · Regime: *%s*", formatINR(opts.SIPAmount), opts.Regime))
	if opts.NiftyPE > 0 {
		sb.WriteString(fmt.Sprintf(" · PE: *%.2f*", opts.NiftyPE))
	}
	sb.WriteString("\n")
	if opts.DebtReserve > 0 {
		sb.WriteString(fmt.Sprintf("Debt reserve: *%s* _(panic buys are extra)_\n", formatINR(opts.DebtReserve)))
	}

	sb.WriteString("\n*From monthly SIP*\n")
	if len(sipBuckets) == 0 {
		sb.WriteString("_No allocations_\n")
	} else {
		for _, b := range sipBuckets {
			sb.WriteString(fmt.Sprintf("\n_%s_ — %s\n", b.title, formatINR(b.total)))
			for _, r := range b.recs {
				sb.WriteString(telegramRecLine(r))
			}
		}
	}

	if len(reserveBuckets) > 0 {
		sb.WriteString("\n*From debt reserve* _(extra)_\n")
		for _, b := range reserveBuckets {
			sb.WriteString(fmt.Sprintf("\n_%s_ — %s\n", b.title, formatINR(b.total)))
			for _, r := range b.recs {
				sb.WriteString(telegramRecLine(r))
			}
		}
	}

	sb.WriteString("\n*Totals*\n")
	sb.WriteString(fmt.Sprintf("• Equity: %s (%.0f%% of SIP)\n", formatINR(sipEquity), pct(sipEquity, opts.SIPAmount)))
	sb.WriteString(fmt.Sprintf("• Debt: %s (%.0f%% of SIP)\n", formatINR(t.sipDebt), pct(t.sipDebt, opts.SIPAmount)))
	if t.reserveTotal > 0 {
		sb.WriteString(fmt.Sprintf("• Reserve extra: %s\n", formatINR(t.reserveTotal)))
		sb.WriteString(fmt.Sprintf("• *Cash needed: %s*\n", formatINR(t.grandTotal)))
	}

	return sb.String()
}

func telegramRecLine(r models.AllocationRecommendation) string {
	label := AssetLabel(r.AssetSymbol)
	var sb strings.Builder
	if label != r.AssetSymbol {
		sb.WriteString(fmt.Sprintf("• *%s* (%s) — %s\n", label, r.AssetSymbol, formatINR(r.Amount)))
	} else {
		sb.WriteString(fmt.Sprintf("• *%s* — %s\n", r.AssetSymbol, formatINR(r.Amount)))
	}
	if r.Reason != "" {
		sb.WriteString(fmt.Sprintf("  _%s_\n", escapeMarkdown(r.Reason)))
	}
	return sb.String()
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer("_", "\\_", "*", "\\*", "[", "\\[", "`", "\\`")
	return replacer.Replace(s)
}
