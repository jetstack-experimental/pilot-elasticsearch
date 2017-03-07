package manager

import "gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"

type kubernetesOptions interface {
	// StatefulSetName returns the name of the StatefulSet
	// containing the clusters data nodes
	StatefulSetName() string
	// PodName returns the name of this pod
	PodName() string
	// Namespace returns the namespace the cluster runs within
	Namespace() string
}

type elasticsearchOptions interface {
	// Role returns the role of this node in the cluster
	Role() util.Role
	// PluginsBin returns the path to the elasticsearch plugins binary
	PluginsBin() string
	// ElasticsearchBin returns the path to the elasticsearch binary
	ElasticsearchBin() string
}

type Options interface {
	kubernetesOptions
	elasticsearchOptions

	// SidecarUsername returns the password for the sidecar elasticsearch account
	SidecarUsername() string
	// SidecarPassword returns the password for the sidecar elasticsearch account
	SidecarPassword() string
}

func NewOptions(fns ...optionsFn) Options {
	opts := &optionsImpl{}
	for _, fn := range fns {
		fn(opts)
	}
	return opts
}

type optionsImpl struct {
	statefulSetName  string
	podName          string
	namespace        string
	role             util.Role
	pluginsBin       string
	elasticsearchBin string
	sidecarUsername  string
	sidecarPassword  string
}

var _ Options = &optionsImpl{}

func (o *optionsImpl) StatefulSetName() string { return o.statefulSetName }
func (o *optionsImpl) PodName() string         { return o.podName }
func (o *optionsImpl) Namespace() string       { return o.namespace }
func (o *optionsImpl) Role() util.Role         { return o.role }
func (o *optionsImpl) PluginsBin() string {
	if len(o.pluginsBin) > 0 {
		return o.pluginsBin
	}
	return "elasticsearch-plugin"
}
func (o *optionsImpl) ElasticsearchBin() string {
	if len(o.elasticsearchBin) > 0 {
		return o.elasticsearchBin
	}
	return "elasticsearch"
}
func (o *optionsImpl) SidecarUsername() string { return o.sidecarUsername }
func (o *optionsImpl) SidecarPassword() string { return o.sidecarPassword }

type optionsFn func(*optionsImpl)

func SetStatefulSetName(s string) optionsFn {
	return func(o *optionsImpl) {
		o.statefulSetName = s
	}
}

func SetPodName(s string) optionsFn {
	return func(o *optionsImpl) {
		o.podName = s
	}
}

func SetNamespace(s string) optionsFn {
	return func(o *optionsImpl) {
		o.namespace = s
	}
}

func SetRole(s util.Role) optionsFn {
	return func(o *optionsImpl) {
		o.role = s
	}
}

func SetPluginsBin(s string) optionsFn {
	return func(o *optionsImpl) {
		o.pluginsBin = s
	}
}

func SetElasticSearchBin(s string) optionsFn {
	return func(o *optionsImpl) {
		o.elasticsearchBin = s
	}
}

func SetSidecarUsername(s string) optionsFn {
	return func(o *optionsImpl) {
		o.sidecarUsername = s
	}
}

func SetSidecarPassword(s string) optionsFn {
	return func(o *optionsImpl) {
		o.sidecarPassword = s
	}
}
