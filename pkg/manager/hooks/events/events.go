package events

type Event string

const (
	// ScaleDownEvent is a StatefulSet scale-down event
	ScaleDownEvent Event = "scaleDown"
)
