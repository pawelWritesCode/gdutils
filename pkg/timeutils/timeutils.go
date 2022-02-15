package timeutils

const (
	// TimeDirectionBackward describes backward time direction
	TimeDirectionBackward TimeDirection = "backward"

	// TimeDirectionForward describes forward time direction
	TimeDirectionForward TimeDirection = "forward"
)

// TimeDirection represents time flow direction.
type TimeDirection string
