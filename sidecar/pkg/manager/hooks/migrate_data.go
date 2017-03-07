package hooks

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/olivere/elastic"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
)

// DrainShards sets the cluster.routing.allocation.exclude._name key to this
// nodes name, in order to begin draining indices from the node. It then blocks
// until the node contains no documents.
func DrainShards(m manager.Interface) error {
	// Only run this hook on data nodes
	if m.Options().Role() != util.RoleData {
		return nil
	}

	shouldRemove, err := nodeShouldBeRemovedFromCluster(m)

	if err != nil {
		// TODO: retry?
		return fmt.Errorf("error determining whether to remove this node from cluster: %s", err.Error())
	}

	if !shouldRemove {
		log.Printf("data migration not needed")
		return nil
	}

	log.Printf("data migration required")

	// exclude this node from being allocated shards
	err = setExcludeAllocation(m, m.Options().PodName())

	if err != nil {
		// TODO: retry?
		return fmt.Errorf("error removing node from cluster: %s", err.Error())
	}

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
	req, err := m.BuildRequest(
		"PUT",
		"/_cluster/settings",
		"",
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

	resp, err := m.ESClient().Do(req)

	if err != nil {
		return fmt.Errorf("error performing request: %s", err.Error())
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("invalid response code '%d' when removing node from cluster", resp.StatusCode)
	}

	return nil
}

// nodeShouldBeRemovedFromCluster returns true if this node is no longer
// going to be serviced by the StatefulSet because of a scale down event
func nodeShouldBeRemovedFromCluster(m manager.Interface) (bool, error) {
	nodeIndex, err := util.NodeIndex(m.Options().PodName())

	if err != nil {
		return false, fmt.Errorf("error parsing node index: %s", err.Error())
	}

	ps, err := m.KubeClient().Apps().StatefulSets(m.Options().Namespace()).Get(m.Options().StatefulSetName())

	if err != nil {
		return false, fmt.Errorf("error getting statefulset: %s", err.Error())
	}

	return nodeIndex >= int(*ps.Spec.Replicas), nil
}

// waitUntilNodeIsEmpty blocks until the node has 0 documents
func waitUntilNodeIsEmpty(m manager.Interface) error {
	for {
		empty, err := nodeIsEmpty(m)

		if err != nil {
			return fmt.Errorf("error waiting for node to be empty: %s", err.Error())
		}

		if empty {
			return nil
		}

		time.Sleep(time.Second * 1)
	}
}

// nodeIsEmpty returns true if this node contains 0 documents
func nodeIsEmpty(m manager.Interface) (bool, error) {
	req, err := m.BuildRequest(
		"GET",
		"/_nodes/stats",
		"",
		nil,
	)

	if err != nil {
		return false, fmt.Errorf("error constructing request: %s", err.Error())
	}

	resp, err := m.ESClient().Do(req)

	if err != nil {
		return false, fmt.Errorf("error getting node stats: %s", err.Error())
	}

	var nodesStatsResponse elastic.NodesStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&nodesStatsResponse)

	if err != nil {
		return false, fmt.Errorf("error decoding response body: %s", err.Error())
	}

	for _, n := range nodesStatsResponse.Nodes {
		if n.Name == m.Options().PodName() {
			return n.Indices.Docs.Count == 0, nil
		}
	}

	return false, nil
}
