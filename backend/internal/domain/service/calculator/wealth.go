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

}

func (w *wealthCalculator) CalculateExistingDebt() {

}
