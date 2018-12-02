package cmds

import (
	"flag"
	"log"
	"os"

	v "github.com/appscode/go/version"
	prom_util "github.com/appscode/kube-mon/prometheus/v1"
	"github.com/appscode/kutil/tools/cli"
	searchlightcheme "github.com/appscode/searchlight/client/clientset/versioned/scheme"
	stashscheme "github.com/appscode/stash/client/clientset/versioned/scheme"
	voyagerscheme "github.com/appscode/voyager/client/clientset/versioned/scheme"
	kubedbscheme "github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	genericapiserver "k8s.io/apiserver/pkg/server"
	_ "k8s.io/client-go/kubernetes/fake"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

func NewCmdKubed(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "kubed",
		Short:             `Kubed by AppsCode - A Kubernetes Cluster Operator Daemon`,
		Long:              `Kubed is a Kubernetes daemon to perform cluster management tasks. For more information, visit here: https://github.com/appscode/kubed/tree/master/docs`,
		DisableAutoGenTag: true,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Printf("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			cli.SendAnalytics(c, v.Version.Version)

			voyagerscheme.AddToScheme(clientsetscheme.Scheme)
			searchlightcheme.AddToScheme(clientsetscheme.Scheme)
			stashscheme.AddToScheme(clientsetscheme.Scheme)
			kubedbscheme.AddToScheme(clientsetscheme.Scheme)
			prom_util.AddToScheme(clientsetscheme.Scheme)
		},
	}
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	// ref: https://github.com/kubernetes/kubernetes/issues/17162#issuecomment-225596212
	flag.CommandLine.Parse([]string{})
	cmd.PersistentFlags().BoolVar(&cli.EnableAnalytics, "enable-analytics", cli.EnableAnalytics, "send usage events to Google Analytics")

	stopCh := genericapiserver.SetupSignalHandler()
	cmd.AddCommand(NewCmdRun(os.Stdout, os.Stderr, stopCh))
	cmd.AddCommand(NewCmdBackup())
	cmd.AddCommand(NewCmdCheck())
	cmd.AddCommand(v.NewCmdVersion())

	return cmd
}
