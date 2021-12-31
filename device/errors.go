package device

import "fmt"

// InvalidLedCountError is an error that is thrown when the led count is not within the allowed range
type InvalidLedCountError struct {
	Count int
	Min   int
	Max   int
}

func (e *InvalidLedCountError) Error() string {
	return fmt.Sprintf("invalid led count error: count:%d min:%d max:%d", e.Count, e.Min, e.Max)
}

// InvalidLedOffsetError is an error that is thrown when the led offset is not within the allowed range
type InvalidLedOffsetError struct {
	Offset int
	Min    int
	Max    int
}

func (e *InvalidLedOffsetError) Error() string {
	return fmt.Sprintf("invalid led offset error: offset:%d min:%d max:%d", e.Offset, e.Min, e.Max)
}
