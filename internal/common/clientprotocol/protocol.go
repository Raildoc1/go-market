package clientprotocol

import "time"

const (
	Null       OrderStatus = ""
	New        OrderStatus = "NEW"
	Invalid    OrderStatus = "INVALID"
	Processing OrderStatus = "PROCESSING"
	Processed  OrderStatus = "PROCESSED"
)

type OrderStatus string

type Order struct {
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    int64       `json:"accrual"`
	UploadedAt time.Time   `json:"uploaded_at"`
}
