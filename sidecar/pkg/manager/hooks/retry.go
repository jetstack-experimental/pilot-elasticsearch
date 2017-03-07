package hooks

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
)

// Retry will retry a hook 'retries' times, waiting 'period' between
// each attempt
func Retry(h manager.Hook, period time.Duration, retries int) manager.Hook {
	return func(m manager.Interface) error {
		var err error
		for i := 0; i < retries; i++ {
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
