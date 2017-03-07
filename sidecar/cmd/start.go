package cmd

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager/hooks"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/probe"
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

			m := manager.NewManager(
				manager.NewOptions(
					manager.SetStatefulSetName(statefulSetName),
					manager.SetPodName(podName),
					manager.SetNamespace(namespace),
					manager.SetRole(util.Role(role.String())),
					manager.SetSidecarUsername(esSidecarUsername),
					manager.SetSidecarPassword(esSidecarPassword),
				),
				kubeClient,
			)

			// Install plugins before Elasticsearch starts
			m.RegisterHooks(manager.PhasePreStart,
				hooks.InstallPlugins(plugins...),
			)

			// Ensure user exists for the lieutenant
			m.RegisterHooks(manager.PhasePostStart,
				hooks.AllowErrors(
					// Only run on data nodes as a shard is required to exist
					// in order to write user data
					hooks.OnlyRoles(
						hooks.Retry(
							hooks.EnsureAccount(esSidecarUsername, esSidecarPassword, "superuser"),
							time.Second*2, // wait 2 seconds between each attempt
							10,            // try 10 times
						),
						util.RoleData,
					),
				),
			)

			// Run DrainShards followed by AcceptShards.
			// TODO: work out a way to run AcceptShards as a postStop hook by talking
			// to the other nodes in cluster
			m.RegisterHooks(manager.PhasePreStop,
				hooks.OnlyRoles(hooks.DrainShards, util.RoleData),
				hooks.OnlyRoles(hooks.AcceptShards, util.RoleData),
			)

			// Start readiness checker
			go (&probe.Listener{
				Port:  12001,
				Check: m.ReadinessCheck(),
			}).Listen()

			// Start liveness checker
			go (&probe.Listener{
				Port:  12000,
				Check: m.LivenessCheck(),
			}).Listen()

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
