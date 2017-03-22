package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
)

// DrainShards sets the cluster.routing.allocation.exclude._name key to this
// nodes name, in order to begin draining indices from the node. It then blocks
// until the node contains no documents.
func DrainShards(m manager.Interface) error {
	log.Infof("draining shards from node...")

	// exclude this node from being allocated shards
	err := setExcludeAllocation(m, m.Options().PodName())

	if err != nil {
		// TODO: retry?
		return fmt.Errorf("error removing node from cluster: %s", err.Error())
	}

	log.Debugf("successfully excluded shard allocation for node '%s'", m.Options().PodName())

	return waitUntilNodeIsEmpty(m)
}

// AcceptShards clears the cluster.routing.allocation.exclude._name key in the
// managers Elasticsearch cluster. This can be used as a postStop hook after
// running the DrainShards hook
func AcceptShards(m manager.Interface) error {
	return setExcludeAllocation(m, "")
}

// setExcludeAllocation sets the cluster.routing.allocation.exclude._name key
func setExcludeAllocation(m manager.Interface, s string) error {
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

// waitUntilNodeIsEmpty blocks until the node has 0 documents
func waitUntilNodeIsEmpty(m manager.Interface) error {
	for {
		log.Debugf("waiting until node is empty...")
		empty, err := nodeIsEmpty(m)

		if err != nil {
			return fmt.Errorf("error waiting for node to be empty: %s", err.Error())
		}

		log.Debugf("node is empty: %t", empty)
		if empty {
			return nil
		}

		time.Sleep(time.Second * 2)
	}
}

// nodeIsEmpty returns true if this node contains 0 documents
func nodeIsEmpty(m manager.Interface) (bool, error) {
	cl, err := m.Client()
	if err != nil {
		return false, err
	}

	resp, err := cl.NodesStats().Do(context.TODO())

	if err != nil {
		return false, fmt.Errorf("error querying node stats: %s", err.Error())
	}

	for _, n := range resp.Nodes {
		if n.Name == m.Options().PodName() {
			log.Debugf("Node '%s' contains %d documents", n.Name, n.Indices.Docs.Count)
			return n.Indices.Docs.Count == 0, nil
		}
	}

	return false, fmt.Errorf("local node not found")
}
