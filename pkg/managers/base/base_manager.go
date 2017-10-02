package base

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers"
)

// Base provides a base Manager implementation. It takes care of hook management &
// firing and generic options handling methods
type base struct {
	// this is a circular reference to the manager implementation
	// that is using this Base as an embedded field. This allows
	// us to call hooks using the top level manager instance's PilotClient.
	// We explicitly avoid maintaining a circular reference to the actual
	// manager to avoid making infinite loops
	client managers.PilotClient

	// options for this manager
	options managers.Options

	phase managers.Phase

	hooks    map[managers.Phase][]managers.Hook
	hookLock sync.RWMutex
}

func New(o managers.Options, cl managers.PilotClient) (managers.Manager, error) {
	return &base{client: cl, options: o}, nil
}

// AddHook registers a hook with this manager for the given Phase
func (b *base) AddHook(p managers.Phase, hooks ...managers.Hook) {
	b.hookLock.Lock()
	defer b.hookLock.Unlock()
	b.hooks[p] = append(b.hooks[p], hooks...)
}

// ExecuteHooks will execute the hooks registered for the given Phase p.
// It will error immediately if one of the hooks fails.
func (b *base) ExecuteHooks(p managers.Phase) error {
	b.hookLock.RLock()
	defer b.hookLock.RUnlock()
	if hooks, ok := b.hooks[p]; ok {
		log.Debugf("executing %d '%s' hooks", len(hooks), p)
		for _, h := range hooks {
			log.Debugf("executing hook...")
			err := h.Execute(b.client)

			if err != nil {
				err = fmt.Errorf("error running hook for phase '%s': %s", p, err.Error())
				log.Warnf(err.Error())
				return err
			}
		}
	}
	return nil
}

// Run starts the Elasticsearch process. It blocks until the process exits.
func (b *base) Run() error {
	return fmt.Errorf("not implemented")
}

// Healthy returns true if the Elasticsearch node this pilot is manager is healthy.
func (b *base) Healthy() bool {
	return false
}

// Version returns the version of the Elasticsearch manager
func (b *base) Version() managers.Version {
	return ""
}

// Phase returns the current phase of the Elasticsearch manager.
func (b *base) Phase() managers.Phase {
	return b.phase
}

func (b *base) PilotClient() managers.PilotClient {
	return b.client
}
