package dto

type TransferRequest struct {
	ToUserID    string `json:"to_user_id" binding:"required"`
	Currency    string `json:"currency" binding:"required,oneof=USD EUR"`
	AmountCents int64  `json:"amount_cents" binding:"required,gt=0"`
}

type ExchangeRequest struct {
	FromCurrency string `json:"from_currency" binding:"required,oneof=USD EUR"`
	AmountCents  int64  `json:"amount_cents" binding:"required,gt=0"`
}

type GetTransactionsRequest struct {
	Type  string `form:"type"`
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
}
