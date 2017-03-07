package hooks

import (
	"fmt"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager/hooks/events"
)

func OnEvent(e events.Event, hs ...manager.Hook) manager.Hook {
	return func(m manager.Interface) error {
		switch e {
		case events.ScaleDownEvent:
			scaleDown, err := events.ScaleDown(
				m.KubeClient(),
				m.Options().Namespace(),
				m.Options().StatefulSetName(),
				m.Options().PodName(),
			)

			if err != nil {
				return fmt.Errorf("error checking scale down event")
			}

			if scaleDown {
				for _, h := range hs {
					if err := h(m); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}
