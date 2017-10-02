package hooks

import (
	"fmt"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager/hooks/events"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers"
)

func OnEvent(e events.Event, hs ...managers.Hook) managers.Hook {
	return managers.NewHook(func(m managers.PilotClient) error {
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
	}, managers.AllVersions...)
}
