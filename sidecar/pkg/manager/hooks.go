package manager

// Hook describes a function that can be used to hook into phase
// change events in Elasticsearch
type Hook func(Interface) error

// Phase describes the phase that the Elasticsearch
// manager is in.
type Phase string

const (
	// PhasePreStart occurs just before the Elasticsearch manager has
	// launched the Elasticsearch process
	PhasePreStart Phase = "preStart"
	// PhasePostStart occurs just after the Elasticsearch manager has
	// launched the Elasticsearch process
	PhasePostStart Phase = "postStart"
	// PhasePreStop occurs just before the Elasticsearch manager stops
	// the Elasticsearch process
	PhasePreStop Phase = "preStop"
	// PhasePostStop occurs just after the Elasticsearch manager has
	// stopped the Elasticsearch process
	PhasePostStop Phase = "postStop"
)
