package utils

import (
	"math"
)

// CalculateAnnuityPayment рассчитывает аннуитетный платеж
// amount - сумма кредита
// annualRate - годовая процентная ставка (%)
// months - срок в месяцах
func CalculateAnnuityPayment(amount float64, annualRate float64, months int) float64 {
	monthlyRate := annualRate / 100 / 12

	if monthlyRate == 0 {
		return amount / float64(months)
	}

	factor := math.Pow(1+monthlyRate, float64(months))
	payment := amount * monthlyRate * factor / (factor - 1)
	return payment
}

// CalculatePaymentSchedule рассчитывает график аннуитетных платежей
func CalculatePaymentSchedule(amount float64, annualRate float64, months int, monthlyPayment float64) []struct {
	Number    int
	Principal float64
	Interest  float64
	Balance   float64
} {
	monthlyRate := annualRate / 100 / 12
	remainingBalance := amount
	schedule := make([]struct {
		Number    int
		Principal float64
		Interest  float64
		Balance   float64
	}, months)

	for i := 0; i < months; i++ {
		interest := remainingBalance * monthlyRate
		principal := monthlyPayment - interest

		if principal > remainingBalance {
			principal = remainingBalance
		}

		remainingBalance -= principal

		schedule[i] = struct {
			Number    int
			Principal float64
			Interest  float64
			Balance   float64
		}{
			Number:    i + 1,
			Principal: principal,
			Interest:  interest,
			Balance:   remainingBalance,
		}
	}

	return schedule
}

// CalculateOverduePenalty рассчитывает пеню за просрочку (+10% к сумме платежа)
func CalculateOverduePenalty(amount float64) float64 {
	return amount * 1.1 // +10%
}
