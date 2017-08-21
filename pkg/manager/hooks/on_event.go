package hooks

import (
	"fmt"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager/hooks/events"
)

func OnEvent(e events.Event, hs ...manager.Hook) manager.Hook {
	return func(m manager.Interface) error {
		switch e {
		case events.ScaleDownEvent:
			scaleDown, err := events.ScaleDown(
				m.KubeClient(),
				m.Options().Namespace(),
				m.Options().ControllerKind(),
				m.Options().ControllerName(),
				m.Options().PodName(),
			)

			if err != nil {
				return fmt.Errorf("error checking scale down event: %s", err.Error())
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
