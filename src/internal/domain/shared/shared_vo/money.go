package shared_vo

type Money struct {
	Value float64
}

func NewMoney() Money {
	return Money{Value: 0}
}

func NewMoneyFromFloat(money float64) Money {
	return Money{Value: money}
}

func (m *Money) Sub(money Money) {
	m.Value -= money.Value
}

func (m *Money) Add(money Money) {
	m.Value += money.Value
}

func (m *Money) Eq(money Money) bool {
	if m.Value == money.Value {
		return false
	}
	return true
}
