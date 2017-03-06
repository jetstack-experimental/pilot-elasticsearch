package hooks

import (
	"fmt"
	"os"
	"os/exec"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/es"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
)

// InstallPlugins returns a hook that will install the plugins specified in
// plugins. If any of the plugins fail to install, an error will be returned
func InstallPlugins(plugins ...string) func(manager.Interface) error {
	return func(m manager.Interface) error {
		for _, plugin := range plugins {
			if plugin == "" {
				continue
			}

			cmd := exec.Command(m.Options().PluginsBin(), "install", plugin)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(), es.Env(m.Options().Role())...)

			err := cmd.Run()

			if err != nil {
				return fmt.Errorf("error installing %s: %s", plugin, err.Error())
			}
		}
		return nil
	}
}
