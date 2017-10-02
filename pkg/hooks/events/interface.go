package events

// Event is a function that returns true if the event that this checks is occurring, or false otherwise.
// For example a 'decommissioned' hook would return true if this pilot had been marked decommissioned.
type Event func() (bool, error)
