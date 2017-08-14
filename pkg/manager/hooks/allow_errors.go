package hooks

import (
	log "github.com/Sirupsen/logrus"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager"
)

// AllowErrors will execute the provided hook, logging
// errors but otherwise ignoring them
func AllowErrors(h manager.Hook) manager.Hook {
	return func(m manager.Interface) error {
		err := h(m)

		if err != nil {
			log.Warnf("skipping error executing hook: %s", err.Error())
		}

		return nil
	}
}
