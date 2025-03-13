package accrualsystemprotocol

import "github.com/shopspring/decimal"

const (
	Registered OrderStatus = "REGISTERED"
	Invalid    OrderStatus = "INVALID"
	Processing OrderStatus = "PROCESSING"
	Processed  OrderStatus = "PROCESSED"
)

type OrderStatus string

type Order struct {
	Number  string          `json:"order"`
	Status  OrderStatus     `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}
