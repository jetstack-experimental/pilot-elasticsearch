package manager

import (
	"os/exec"

	"github.com/olivere/elastic"
	"k8s.io/client-go/kubernetes"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers/base"
)

const (
	pluginBinary = "elasticsearch-plugin"
)

type manager struct {
	managers.Manager

	options    managers.Options
	esClient   *elastic.Client
	kubeClient *kubernetes.Clientset
	phase      managers.Phase
	esCmd      *exec.Cmd
}

var _ managers.Manager = (*manager)(nil)

func (m *manager) Run() error {
	return nil
}

func (m *manager) Healthy() bool {
	return true
}

func (m *manager) Version() managers.Version {
	return managers.Version5
}

func (m *manager) PilotClient() managers.PilotClient {
	return m
}

func init() {
	managers.Register(managers.Version5, func(opts managers.Options) (managers.Manager, error) {
		m := &manager{options: opts}
		b, err := base.New(opts, m.PilotClient())
		if err != nil {
			return nil, err
		}
		m.Manager = b
		return m, nil
	})
}
