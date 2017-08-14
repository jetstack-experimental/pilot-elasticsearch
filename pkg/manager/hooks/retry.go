package hooks

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager"
)

// Retry will retry a hook 'retries' times, waiting 'period' between
// each attempt
func Retry(h manager.Hook, period time.Duration, retries int) manager.Hook {
	return func(m manager.Interface) error {
		var err error
		for i := 0; i < retries; i++ {
			log.Debugf("attempting to executing hook...")
			err = h(m)

			if err != nil {
				log.Warnf("skipping error executing hook: %s", err.Error())
				time.Sleep(period)
				continue
			}

			break
		}

		if err != nil {
			return fmt.Errorf("error executing hook: %s", err.Error())
		}

		return nil
	}
}
