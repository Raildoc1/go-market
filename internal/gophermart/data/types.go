package data

type Status string

const (
	NewStatus        = Status("NEW")
	ProcessingStatus = Status("PROCESSING")
	ProcessedStatus  = Status("PROCESSED")
	InvalidStatus    = Status("INVALID")
)
