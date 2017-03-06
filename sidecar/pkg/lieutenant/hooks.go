package lieutenant

type Hook func(Interface) error

type Phase string

const (
	// PhasePreStart occurs after the Elasticsearch manager has been
	// constructed, but before Elasticsearch itself has been started on
	// this node.
	PhasePreStart Phase = "preStart"
	// PhasePostStart occurs just after the Elasticsearch manager has
	// launched the Elasticsearch process
	PhasePostStart Phase = "postStart"
	PhasePreStop   Phase = "preStop"
	PhasePostStop  Phase = "postStop"
)
