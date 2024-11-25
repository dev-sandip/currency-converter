package currency

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type ExchangeRates struct {
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}

// Default exchange rates to use when API is not available
var defaultRates = ExchangeRates{
	Timestamp: time.Now().Unix(),
	Base:      "USD",
	Rates: map[string]float64{
		"EUR": 0.92,
		"GBP": 0.79,
		"JPY": 151.37,
		"AUD": 1.53,
		"CAD": 1.36,
		"CHF": 0.90,
		"CNY": 7.24,
		"INR": 83.31,
		"NZD": 1.65,
		"BRL": 5.04,
	},
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
}

func ReadData() (*ExchangeRates, error) {
	file, err := os.Open("data.json")
	if os.IsNotExist(err) {
		// Create default data if file doesn't exist
		err = writeDataStruct(&defaultRates)
		if err != nil {
			return nil, fmt.Errorf("error creating data.json: %v", err)
		}
		return &defaultRates, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var data ExchangeRates
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func writeData(data []byte) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func writeDataStruct(data *ExchangeRates) error {
	file, err := os.Create("data.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func fetchData(apiKey string) ([]byte, error) {
	url := "https://openexchangerates.org/api/latest.json?app_id=" + apiKey
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return io.ReadAll(response.Body)
}

func UpdateRates() (*ExchangeRates, error) {
	loadEnv()
	apiKey := os.Getenv("CURRENCY_API")
	if apiKey == "" {
		fmt.Println("API key not found in .env, using default rates")
		return &defaultRates, nil
	}

	data, err := ReadData()
	if err != nil {
		return nil, fmt.Errorf("error reading data: %v", err)
	}

	currentTime := time.Now().Unix()
	if currentTime-data.Timestamp < 10*60*60 {
		fmt.Println("Returning cached data from data.json")
		return data, nil
	}

	newData, err := fetchData(apiKey)
	if err != nil {
		fmt.Println("Error fetching data from API, using cached data")
		return data, nil
	}

	var rates ExchangeRates
	err = json.Unmarshal(newData, &rates)
	if err != nil {
		return nil, fmt.Errorf("error parsing API response: %v", err)
	}

	err = writeDataStruct(&rates)
	if err != nil {
		return nil, fmt.Errorf("error writing data to file: %v", err)
	}

	fmt.Println("New data fetched and written to data.json!")
	return &rates, nil
}

func GetAvailableCurrencies() []string {
	data, err := ReadData()
	if err != nil {
		return getDefaultCurrencies()
	}

	currencies := make([]string, 0, len(data.Rates)+1)
	currencies = append(currencies, data.Base) // Add base currency
	for currency := range data.Rates {
		currencies = append(currencies, currency)
	}
	return currencies
}

func getDefaultCurrencies() []string {
	currencies := make([]string, 0, len(defaultRates.Rates)+1)
	currencies = append(currencies, defaultRates.Base)
	for currency := range defaultRates.Rates {
		currencies = append(currencies, currency)
	}
	return currencies
}
