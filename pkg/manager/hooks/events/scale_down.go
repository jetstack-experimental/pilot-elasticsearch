package events

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/util"
)

// ScaleDown will return true if this pod is shutting down as part of a
// StatefulSet scale down event.
func ScaleDown(cl *kubernetes.Clientset, namespace, controllerKind, controllerName, podName string) (bool, error) {
	nodeIndex, err := util.NodeIndex(podName)

	if err != nil {
		return false, fmt.Errorf("error parsing node index: %s", err.Error())
	}

	var desiredReplicas int
	switch controllerKind {
	case "StatefulSet":
		ps, err := cl.Apps().StatefulSets(namespace).Get(controllerName, metav1.GetOptions{})

		if err != nil {
			return false, fmt.Errorf("error getting statefulset '%s': %s", controllerName, err.Error())
		}

		desiredReplicas = int(*ps.Spec.Replicas)
	case "Deployment":
		ps, err := cl.Extensions().Deployments(namespace).Get(controllerName, metav1.GetOptions{})

		if err != nil {
			return false, fmt.Errorf("error getting deployment '%s': %s", controllerName, err.Error())
		}

		desiredReplicas = int(*ps.Spec.Replicas)
	default:
		return false, fmt.Errorf("invalid controller kind specified: '%s'", controllerKind)
	}

	return nodeIndex >= desiredReplicas, nil
}
