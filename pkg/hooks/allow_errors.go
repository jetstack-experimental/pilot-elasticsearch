package hooks

import (
	log "github.com/Sirupsen/logrus"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers"
)

// AllowErrors will execute the provided hook, logging
// errors but otherwise ignoring them
func AllowErrors(h managers.Hook) managers.Hook {
	return managers.NewHook(func(m managers.PilotClient) error {
		err := h.Execute(m)

		if err != nil {
			log.Warnf("skipping error executing hook: %s", err.Error())
		}

		return nil
	}, managers.AllVersions...)
}
