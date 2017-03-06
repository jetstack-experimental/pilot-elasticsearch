package lieutenant

import (
	"fmt"
	"os"
	"syscall"
)

func (m *Manager) Shutdown() error {
	defer os.Exit(1)
	if err := m.transitionPhase(PhasePreStop); err != nil {
		return fmt.Errorf("error running lieutenant pre stop hooks: %s", err.Error())
	}

	if m.esCmd != nil {
		m.esCmd.Process.Signal(syscall.SIGTERM)
		state, err := m.esCmd.Process.Wait()

		// we'll skip this error so that postStop hooks are fired
		if err != nil {
			return fmt.Errorf("elasticsearch exited with error: %s", err.Error())
		}

		if !state.Exited() {
			return fmt.Errorf("warning: elasticsearch has not exited...")
		}
	}

	if err := m.transitionPhase(PhasePostStop); err != nil {
		return fmt.Errorf("error running lieutenant post stop hooks: %s", err.Error())
	}

	return nil
}
