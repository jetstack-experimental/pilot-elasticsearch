package manager

import (
	"fmt"

	"net/url"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/util"
)

type kubernetesOptions interface {
	// ControllerKind returns the kind of the controller
	// managing this node
	ControllerKind() string
	// ControllerName returns the name of the controller
	// managing this node
	ControllerName() string
	// PodName returns the name of this pod
	PodName() string
	// Namespace returns the namespace the cluster runs within
	Namespace() string
}

type elasticsearchOptions interface {
	// Role returns the role of this node in the cluster
	Roles() []util.Role
	// PluginsBin returns the path to the elasticsearch plugins binary
	PluginsBin() string
	// ElasticsearchBin returns the path to the elasticsearch binary
	ElasticsearchBin() string
	// ClusterURL that can be used to talk to client nodes
	ClusterURL() url.URL
}

type Options interface {
	kubernetesOptions
	elasticsearchOptions

	// SidecarUsername returns the password for the sidecar elasticsearch account
	SidecarUsername() string
	// SidecarPassword returns the password for the sidecar elasticsearch account
	SidecarPassword() string
}

func NewOptions(fns ...optionsFn) (Options, error) {
	opts := &optionsImpl{}
	for _, fn := range fns {
		if err := fn(opts); err != nil {
			return nil, err
		}
	}
	return opts, nil
}

type optionsImpl struct {
	controllerKind, controllerName string
	podName                        string
	namespace                      string
	roles                          []util.Role
	pluginsBin                     string
	elasticsearchBin               string
	sidecarUsername                string
	sidecarPassword                string
	clusterURL                     url.URL
}

var _ Options = &optionsImpl{}

func (o *optionsImpl) ControllerKind() string { return o.controllerKind }
func (o *optionsImpl) ControllerName() string { return o.controllerName }
func (o *optionsImpl) PodName() string        { return o.podName }
func (o *optionsImpl) Namespace() string      { return o.namespace }
func (o *optionsImpl) Roles() []util.Role     { return o.roles }
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
func (o *optionsImpl) ClusterURL() url.URL     { return o.clusterURL }

type optionsFn func(*optionsImpl) error

func SetControllerKind(s string) optionsFn {
	return func(o *optionsImpl) error {
		switch s {
		case "Deployment", "StatefulSet":
			break
		default:
			return fmt.Errorf("invalid controller kind '%s'. must be one of 'StatefulSet', 'Deployment'", s)
		}
		o.controllerKind = s
		return nil
	}
}

func SetControllerName(s string) optionsFn {
	return func(o *optionsImpl) error {
		o.controllerName = s
		return nil
	}
}

func SetPodName(s string) optionsFn {
	return func(o *optionsImpl) error {
		o.podName = s
		return nil
	}
}

func SetNamespace(s string) optionsFn {
	return func(o *optionsImpl) error {
		o.namespace = s
		return nil
	}
}

func SetRoles(s []util.Role) optionsFn {
	return func(o *optionsImpl) error {
		o.roles = s
		return nil
	}
}

func SetPluginsBin(s string) optionsFn {
	return func(o *optionsImpl) error {
		o.pluginsBin = s
		return nil
	}
}

func SetElasticSearchBin(s string) optionsFn {
	return func(o *optionsImpl) error {
		o.elasticsearchBin = s
		return nil
	}
}

func SetSidecarUsername(s string) optionsFn {
	return func(o *optionsImpl) error {
		o.sidecarUsername = s
		return nil
	}
}

func SetSidecarPassword(s string) optionsFn {
	return func(o *optionsImpl) error {
		o.sidecarPassword = s
		return nil
	}
}

func SetClusterURL(s string) optionsFn {
	return func(o *optionsImpl) error {
		if url, err := url.Parse(s); err != nil {
			return err
		} else {
			o.clusterURL = *url
			return nil
		}
	}
}
