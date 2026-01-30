package models

const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
)

const (
	TransactionTypeTransfer        = "transfer"
	TransactionTypeExchange        = "exchange"
	TransactionTypeInitialDeposit  = "initial_deposit"
)


const (
	ExchangeRateUSDtoEURNum   int64 = 23
	ExchangeRateUSDtoEURDenom int64 = 25

	ExchangeRateEURtoUSDNum   int64 = 25
	ExchangeRateEURtoUSDDenom int64 = 23
)

const (
	MinExchangeAmountCents int64 = 10
)

