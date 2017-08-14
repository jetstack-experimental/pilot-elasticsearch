package hooks

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jetstack-experimental/navigator/pkg/apis/marshal/v1alpha1"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/es"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager"
)

// InstallPlugins returns a hook that will install the plugins specified in
// plugins. If any of the plugins fail to install, an error will be returned
func InstallPlugins(plugins ...v1alpha1.ElasticsearchClusterPlugin) func(manager.Interface) error {
	return func(m manager.Interface) error {
		for _, plugin := range plugins {
			if plugin.Name == "" {
				continue
			}

			cmd := exec.Command(m.Options().PluginsBin(), "install", plugin.Name)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(), es.Env(m.Options().Roles())...)

			err := cmd.Run()

			if err != nil {
				return fmt.Errorf("error installing %s: %s", plugin, err.Error())
			}
		}
		return nil
	}
}
