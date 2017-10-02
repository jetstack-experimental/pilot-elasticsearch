package manager

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/es"
)

func (m *manager) InstallPlugin(p v1alpha1.ElasticsearchClusterPlugin) error {
	cmd := exec.Command(m.options.PluginsBinary, "install", p.Name)
	cmd.Env = append(os.Environ(), es.Env(m.options.Roles)...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error installing plugin '%s': %s", p.Name, err.Error())
	}

	return nil
}

func (m *manager) DrainNode(s string) error {
	log.Debugf("excluding shard allocation for node '%s'", s)
	req, err := m.BuildRequest(
		"PUT",
		"/_cluster/settings",
		"",
		true,
		strings.NewReader(
			fmt.Sprintf(`
			{
				"transient": {
					"cluster.routing.allocation.exclude._name": "%s"
				}	
			}`, s),
		),
	)

	if err != nil {
		return fmt.Errorf("error constructing request: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.ESClient().Do(req)

	if err != nil {
		return fmt.Errorf("error performing request: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("invalid response code '%d' when removing node from cluster", resp.StatusCode)
	}

	return nil
}

func (m *manager) SetEnvironment(string, string) {
}
