package events

import (
	"fmt"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
	"k8s.io/client-go/kubernetes"
)

// ScaleDown will return true if this pod is shutting down as part of a
// StatefulSet scale down event.
func ScaleDown(cl *kubernetes.Clientset, namespace, statefulSetName, podName string) (bool, error) {
	nodeIndex, err := util.NodeIndex(podName)

	if err != nil {
		return false, fmt.Errorf("error parsing node index: %s", err.Error())
	}

	ps, err := cl.Apps().StatefulSets(namespace).Get(statefulSetName)

	if err != nil {
		return false, fmt.Errorf("error getting statefulset: %s", err.Error())
	}

	return nodeIndex >= int(*ps.Spec.Replicas), nil
}
