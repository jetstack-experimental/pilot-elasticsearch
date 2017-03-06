package hooks

import (
	"fmt"
	"os"
	"os/exec"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/es"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/lieutenant"
)

func InstallPlugins(plugins ...string) func(lieutenant.Interface) error {
	return func(m lieutenant.Interface) error {
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
