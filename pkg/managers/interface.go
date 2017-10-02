package managers

import (
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
)

// Manager is an interface for managing an Elasticsearch process
type Manager interface {
	// Run starts the Elasticsearch process. It blocks until the process exits.
	Run() error
	// AddHook adds the given hooks to the given phase
	AddHook(Phase, ...Hook)
	// ExecuteHooks executes the hooks for the given phase
	ExecuteHooks(Phase) error
	// Phase returns the current phase of the Elasticsearch manager.
	Phase() Phase
	// Healthy returns true if the Elasticsearch node this pilot is manager is healthy.
	Healthy() bool
	// Version returns the version of the Elasticsearch manager
	Version() Version
	// PilotClient returns a PilotClient that can be used to perform common cluster
	// actions using this Pilot
	PilotClient() PilotClient
}

// PilotClient is a manager helper that exposes common methods that should
// be possible regardless of Elasticsearch version. A Manager should provide
// a PilotClient that can be used for performing actions on the Elasticsearch
// cluster.
type PilotClient interface {
	// InstallPlugin will install a plugin for the given Elasticsearch process
	// to use. This is not guaranteed to work correctly if called after the
	// Elasticsearch process has transitioned from the preStart phase.
	InstallPlugin(v1alpha1.ElasticsearchClusterPlugin) error

	// DrainNode will exclude the node with the given ID from being allocated shards
	DrainNode(nodeID string) error

	// SetEnvironment sets an environment variable for the Elasticsearch process.
	// This must be called before the manager transitions from the preStart phase.
	SetEnvironment(key, val string)

	// Roles returns the roles assigned to this node
	Roles() []Role
}

// Hook is an interface for defining lifecycle hooks for
// the Elasticsearch manager
type Hook interface {
	// Execute runs this hook
	Execute(PilotClient) error
	// Supported returns true if this hook is supports the given manager Version
	Supported(Version) bool
}

// Version represents an Elasticsearch version
type Version string

const (
	// Version5 of Elasticsearch
	Version5 Version = "v5"
	// Version2 of Elasticsearch
	Version2 Version = "v2"
	// Version1 of Elasticsearch
	Version1 Version = "v1"
	// VersionBase implements functionality
	// common across all Elasticsearch versions
	VersionBase Version = "base"
)

var (
	// AllVersions are all Elasticsearch versions known by this pilot
	AllVersions = []Version{Version1, Version2, Version5}
)
