package managers

type Options struct {
	// Roles is the roles of this node in the cluster
	Roles []Role
	// PluginsBinary is the path to the elasticsearch plugins binary
	PluginsBinary string
	// ElasticsearchBinary is the path to the elasticsearch binary
	ElasticsearchBinary string
	// ClusterURL that can be used to communicate with the client nodes
	ClusterURL string
}
