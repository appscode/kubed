package framework

import (
	"log"
	"strings"
	"flag"

	"path/filepath"
	"k8s.io/client-go/util/homedir"
	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/flags"
)

type E2EConfig struct {
	Master            string
	KubeConfig        string
	CloudProviderName string
	HAProxyImageName  string
	TestNamespace     string
	// IngressClass      string
	InCluster         bool
	Cleanup           bool
	DaemonHostName    string
	LBPersistIP       string
	// RBACEnabled       bool
	// TestCertificate   bool
}

func init() {
	/*fmt.Println("******************Hello Init", flag.String("master", "", ""))
	flag.StringVar(&testConfigs.Master, "master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	flag.StringVar(&testConfigs.KubeConfig, "kubeconfig", "", "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&testConfigs.CloudProviderName, "cloud-provider", "", "Name of cloud provider")
	flag.StringVar(&testConfigs.HAProxyImageName, "haproxy-image", "appscode/haproxy:1.7.9-4.0.0-rc.1", "haproxy image name to be run")
	flag.StringVar(&testConfigs.TestNamespace, "namespace", "test-"+rand.Characters(5), "Run tests in this namespaces")
	// &testConfigs.Master = ""
	&testConfigs.KubeConfig = ""*/

	enableLogging()
}


var testConfigs E2EConfig

func enableLogging() {
	flag.Set("logtostderr", "true")
	logLevelFlag := flag.Lookup("v")
	if logLevelFlag != nil {
		if len(logLevelFlag.Value.String()) > 0 && logLevelFlag.Value.String() != "0" {
			return
		}
	}
	flags.SetLogLevel(2)
}

func (c *E2EConfig) validate()  {
	/*if c.CloudProviderName == "" {
		log.Fatal("Provider name required, not provided")
	}*/

	if len(c.KubeConfig) == 0 {
		c.KubeConfig = filepath.Join(homedir.HomeDir(), ".kube/config")
	}

	if len(c.TestNamespace) == 0 {
		c.TestNamespace = rand.WithUniqSuffix("test-kubed")
	}

	if !strings.HasPrefix(c.TestNamespace, "test-") {
		log.Fatal("Namespace is not a Test namespace")
	}
}
