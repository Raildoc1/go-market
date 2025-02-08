package data

type Status string

const (
	NullStatus       = Status("")
	NewStatus        = Status("NEW")
	ProcessingStatus = Status("PROCESSING")
	ProcessedStatus  = Status("PROCESSED")
	InvalidStatus    = Status("INVALID")
)
