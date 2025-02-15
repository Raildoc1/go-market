package data

import (
	"time"

	"github.com/shopspring/decimal"
)

type Status string

const (
	NullStatus       = Status("")
	NewStatus        = Status("NEW")
	ProcessingStatus = Status("PROCESSING")
	ProcessedStatus  = Status("PROCESSED")
	InvalidStatus    = Status("INVALID")
)

type Order struct {
	UploadTime  time.Time
	OrderNumber string
	Accrual     decimal.Decimal
	Status      Status
	UserID      int
}

type Withdrawal struct {
	ProcessTime time.Time
	OrderNumber string
	Amount      decimal.Decimal
	UserID      int
}
