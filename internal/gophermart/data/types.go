package data

import (
	"github.com/shopspring/decimal"
	"time"
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
	OrderNumber string
	UserID      int
	Accrual     decimal.Decimal
	Status      Status
	UploadTime  time.Time
}

type Withdrawal struct {
	OrderNumber string
	UserID      int
	Amount      decimal.Decimal
	ProcessTime time.Time
}
