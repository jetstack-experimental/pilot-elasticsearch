package manager

import (
	"fmt"
	"io/ioutil"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/probe"
)

// Check the health of this Elasticsearch node
func localNodeHealth(m Interface) func() error {
	return func() error {
		req, err := m.BuildRequest("GET", "/_cluster/health", "local=true", true, nil)

		if err != nil {
			return fmt.Errorf("error building health check request: %s", err.Error())
		}

		resp, err := m.ESClient().Do(req)

		if err != nil {
			return fmt.Errorf("error making health check request: %s", err.Error())
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var body string
			if b, err := ioutil.ReadAll(resp.Body); err == nil {
				body = string(b)
			}
			return fmt.Errorf("node unhealthy (%d): %s", resp.StatusCode, body)
		}

		return nil
	}
}

func (m *Manager) ReadinessCheck() probe.Check {
	return probe.CombineChecks(
		// Check the Elasticsearch phase
		func() error {
			switch m.Phase() {
			case PhasePreStart, PhasePreStop, PhasePostStop:
				return fmt.Errorf("elasticsearch not running. phase: %s", m.Phase())
			}
			return nil
		},
		// Check the health of this Elasticsearch node
		localNodeHealth(m),
	)
}

func (m *Manager) LivenessCheck() probe.Check {
	return probe.CombineChecks()
}
