package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	currency "github.com/dev-sandip/currency-converter/api"
)

type ExchangeRates struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

func formatAmount(amount float64, currencyCode string) string {
	var decimals int
	// Currencies typically without decimals
	noDecimalCurrencies := map[string]bool{
		"JPY": true, "KRW": true, "VND": true, "IDR": true,
		"CLP": true, "ISK": true, "HUF": true,
	}

	// Currencies with 3 decimal places
	threeDecimalCurrencies := map[string]bool{
		"BHD": true, "IQD": true, "KWD": true, "OMR": true,
	}

	switch {
	case noDecimalCurrencies[currencyCode]:
		decimals = 0
	case threeDecimalCurrencies[currencyCode]:
		decimals = 3
	default:
		decimals = 2
	}

	// Format the number with the appropriate decimals
	format := fmt.Sprintf("%%.%df", decimals)
	value := fmt.Sprintf(format, amount)

	// Add thousand separators
	parts := strings.Split(value, ".")
	intPart := parts[0]
	var formatted []byte
	for i := len(intPart) - 1; i >= 0; i-- {
		if len(formatted) > 0 && (len(intPart)-i-1)%3 == 0 {
			formatted = append([]byte{','}, formatted...)
		}
		formatted = append([]byte{intPart[i]}, formatted...)
	}

	if len(parts) > 1 {
		return string(formatted) + "." + parts[1]
	}
	return string(formatted)
}

func convertCurrency(amount float64, from, to string, rates *currency.ExchangeRates) (float64, error) {
	// Handle same currency conversion
	if from == to {
		return amount, nil
	}

	// Since USD is the base currency
	// Converting from USD to another currency
	if from == "USD" {
		if rate, exists := rates.Rates[to]; exists {
			return amount * rate, nil
		}
		return 0, fmt.Errorf("target currency %s not found", to)
	}

	// Converting to USD
	if to == "USD" {
		if rate, exists := rates.Rates[from]; exists {
			return amount / rate, nil
		}
		return 0, fmt.Errorf("source currency %s not found", from)
	}

	// Converting between two non-USD currencies
	fromRate, fromExists := rates.Rates[from]
	toRate, toExists := rates.Rates[to]

	if !fromExists || !toExists {
		return 0, fmt.Errorf("currency rates not found")
	}

	// Convert through USD
	amountInUSD := amount / fromRate
	return amountInUSD * toRate, nil
}

func validateAmount(amount string) error {
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("please enter a valid number")
	}
	if val <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	return nil
}

func main() {
	var (
		selectedCurrency string
		amount           string
		targetCurrency   string
	)

	// Get available currencies
	rates, err := currency.UpdateRates()
	if err != nil {
		log.Printf("Warning: Error updating rates: %v", err)
	}

	// Create currency options
	currencies := currency.GetAvailableCurrencies()
	currencyOptions := make([]huh.Option[string], 0, len(currencies))
	for _, curr := range currencies {
		currencyOptions = append(currencyOptions, huh.NewOption(curr, curr))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select source currency").
				Options(currencyOptions...).
				Value(&selectedCurrency),

			huh.NewInput().
				Title("Enter amount").
				Value(&amount).
				Validate(validateAmount),

			huh.NewSelect[string]().
				Title("Select target currency").
				Options(currencyOptions...).
				Value(&targetCurrency),
		),
	)

	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Convert amount string to float
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		log.Fatal("Error parsing amount:", err)
	}

	// Perform the conversion
	convertedAmount, err := convertCurrency(amountFloat, selectedCurrency, targetCurrency, rates)
	if err != nil {
		log.Fatal("Error converting currency:", err)
	}

	// Calculate exchange rate
	rate := convertedAmount / amountFloat

	// Format output
	fmt.Println("\nConversion Result:")
	fmt.Printf("%s %s = %s %s\n",
		selectedCurrency,
		formatAmount(amountFloat, selectedCurrency),
		targetCurrency,
		formatAmount(convertedAmount, targetCurrency))

	// Show exchange rate
	fmt.Printf("Exchange Rate: 1 %s = %s %s\n",
		selectedCurrency,
		formatAmount(rate, targetCurrency),
		targetCurrency)

	// Show equivalent in USD if neither currency is USD
	if selectedCurrency != "USD" && targetCurrency != "USD" {
		usdAmount, _ := convertCurrency(amountFloat, selectedCurrency, "USD", rates)
		fmt.Printf("USD Equivalent: $%s\n", formatAmount(usdAmount, "USD"))
	}
}
