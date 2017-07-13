package controller

import (
	"sync"
	"time"

	"github.com/appscode/kubed/pkg/config"
	"github.com/appscode/kubed/pkg/indexers"
	"github.com/appscode/kubed/pkg/recover"
	srch_cs "github.com/appscode/searchlight/client/clientset"
	scs "github.com/appscode/stash/client/clientset"
	vcs "github.com/appscode/voyager/client/clientset"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	kcs "github.com/k8sdb/apimachinery/client/clientset"
	clientset "k8s.io/client-go/kubernetes"
)

type Options struct {
	Master             string
	KubeConfig         string
	EnableAnalytics    bool
	Indexer            string
	EnableReverseIndex bool
	ServerAddress      string
	ConfigPath         string
}

type Controller struct {
	KubeClient        clientset.Interface
	VoyagerClient     vcs.ExtensionInterface
	SearchlightClient srch_cs.ExtensionInterface
	StashClient       scs.ExtensionInterface
	PromClient        pcm.MonitoringV1alpha1Interface
	KubeDBClient      kcs.ExtensionInterface

	Opt          Options
	Config       config.ClusterConfig
	Saver        *recover.RecoverStuff
	Indexer      *indexers.ResourceIndexer
	ReverseIndex *indexers.ReverseIndexer
	SyncPeriod   time.Duration
	sync.Mutex
}

func (c *Controller) Run() {
	go c.WatchAlertmanagers()
	go c.WatchClusterAlerts()
	go c.WatchConfigMaps()
	go c.WatchDaemonSets()
	go c.WatchDeploymentApps()
	go c.WatchDeploymentExtensions()
	go c.WatchDormantDatabases()
	go c.WatchElastics()
	go c.WatchEvents()
	go c.WatchIngresss()
	go c.WatchJobs()
	go c.watchNamespaces()
	go c.WatchNodeAlerts()
	go c.WatchPersistentVolumeClaims()
	go c.WatchPersistentVolumes()
	go c.WatchPodAlerts()
	go c.WatchPostgreses()
	go c.WatchPrometheuss()
	go c.WatchReplicaSets()
	go c.WatchReplicationControllers()
	go c.WatchRestics()
	go c.WatchSecrets()
	go c.watchService()
	go c.WatchServiceMonitors()
	go c.WatchStatefulSets()
	go c.WatchStorageClasss()
	go c.WatchVoyagerCertificates()
	go c.WatchVoyagerIngresses()
}
