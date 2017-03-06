package hooks

import (
	"fmt"
	"log"
	"strings"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/lieutenant"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
)

func MigrateData(m lieutenant.Interface) error {
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

	err = excludeNodeFromBeingAllocatedShards(m)

	if err != nil {
		// TODO: retry?
		return fmt.Errorf("error removing node from cluster: %s", err.Error())
	}

	return nil
	// TODO: here, wait for the document count for this node to drop to 0
}

func nodeShouldBeRemovedFromCluster(m lieutenant.Interface) (bool, error) {
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

func excludeNodeFromBeingAllocatedShards(m lieutenant.Interface) error {
	log.Printf("Excluding node from cluster")

	req, err := m.BuildRequest(
		"PUT",
		"/_cluster/settings",
		strings.NewReader(
			fmt.Sprintf(`
			{
				"transient": {
					"cluster.routing.allocation.exclude._name": "%s"
				}	
			}`, m.Options().PodName()),
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
