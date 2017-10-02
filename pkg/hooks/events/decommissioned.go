package events

import "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset"

func NewDecommissioned(navClient *clientset.Interface, pilotName string, namespace string) Event {
	return func() (bool, error) {

	}
}
