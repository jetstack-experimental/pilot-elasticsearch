package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/lieutenant"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/lieutenant/hooks"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
)

var (
	apiServerHost   string
	namespace       string
	podName         string
	role            = &util.RoleVar{}
	plugins         []string
	statefulSetName string

	esSidecarUsername = "_sidecar"
	esSidecarPassword string

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "starts the elasticsearch lieutenant",
		Run: func(cmd *cobra.Command, args []string) {
			kubeClient, err := util.NewKubernetesClient(apiServerHost)

			if err != nil {
				log.Fatalf("error creating kubernetes client: %s", err.Error())
			}

			m := lieutenant.NewManager(
				lieutenant.NewOptions(
					lieutenant.SetStatefulSetName(statefulSetName),
					lieutenant.SetPodName(podName),
					lieutenant.SetNamespace(namespace),
					lieutenant.SetRole(util.Role(role.String())),
					lieutenant.SetSidecarUsername(esSidecarUsername),
					lieutenant.SetSidecarPassword(esSidecarPassword),
				),
				kubeClient,
			)

			m.RegisterHook(lieutenant.PhasePreStart, hooks.InstallPlugins(plugins...))
			m.RegisterHook(lieutenant.PhasePostStart, hooks.CreateAdminAccount)
			m.RegisterHook(lieutenant.PhasePreStop, hooks.MigrateData)

			if err := m.Run(); err != nil {
				log.Fatalf("error running elasticsearch: %s", err.Error())
			}
		},
	}
)

func init() {
	// StartCmd flags
	startCmd.PersistentFlags().StringVar(&podName, "podName", "", "The name of this pod")
	startCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "The namespace this node is running in")
	startCmd.PersistentFlags().StringVar(&apiServerHost, "apiServerHost", "", "Kubernetes apiserver host address (overrides autodetection)")
	startCmd.PersistentFlags().StringSliceVarP(&plugins, "plugins", "p", []string{}, "List of Elasticsearch plugins to install")
	startCmd.PersistentFlags().VarP(role, "role", "r", "The role of this Elasticsearch node")
	startCmd.PersistentFlags().StringVar(&statefulSetName, "statefulSetName", "", "Name of the StatefulSet managing data nodes")
	startCmd.PersistentFlags().StringVar(&esSidecarPassword, "sidecarUserPassword", "insecure", "The password to use for the sidecars ElasticSearch user account")

	rootCmd.AddCommand(startCmd)
}
