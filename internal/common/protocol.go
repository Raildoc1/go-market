package common

const (
	Registered OrderStatus = "REGISTERED"
	Invalid    OrderStatus = "INVALID"
	Processing OrderStatus = "PROCESSING"
	Processed  OrderStatus = "PROCESSED"
)

type OrderStatus string

type Order struct {
	Number string      `json:"order"`
	Status OrderStatus `json:"status"`
	Points int64       `json:"accrual"`
}
