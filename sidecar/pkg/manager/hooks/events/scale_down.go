package events

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
)

// ScaleDown will return true if this pod is shutting down as part of a
// StatefulSet scale down event.
func ScaleDown(cl *kubernetes.Clientset, namespace, statefulSetName, podName string) (bool, error) {
	nodeIndex, err := util.NodeIndex(podName)

	if err != nil {
		return false, fmt.Errorf("error parsing node index: %s", err.Error())
	}

	ps, err := cl.Apps().StatefulSets(namespace).Get(statefulSetName, metav1.GetOptions{})

	if err != nil {
		return false, fmt.Errorf("error getting statefulset: %s", err.Error())
	}

	return nodeIndex >= int(*ps.Spec.Replicas), nil
}
