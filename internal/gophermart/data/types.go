package data

import (
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
	UserId      int
	Accrual     int64
	Status      Status
	UploadTime  time.Time
}
