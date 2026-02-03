package calculator

type WealthCalculator interface {
	CalculateExistingWealth()
	CalculateExistingDebt()
}

type wealthCalculator struct {
}

func NewWealthCalculator() WealthCalculator {
	return &wealthCalculator{}
}

func (w *wealthCalculator) CalculateExistingWealth() {
	// equity + debt

}

func (w *wealthCalculator) CalculateExistingDebt() {
	// only debt
}
