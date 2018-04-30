package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/kubecon/cmd/session-controller/controller"
	clientset "k8s.io/kubecon/pkg/client/clientset/versioned"
	informers "k8s.io/kubecon/pkg/client/informers/externalversions"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeconClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeconInformerFactory := informers.NewSharedInformerFactory(kubeconClient, time.Second*30)

	controller := controller.NewController(kubeconClient, kubeconInformerFactory)

	go kubeconInformerFactory.Start(wait.NeverStop)

	if err = controller.Run(2, wait.NeverStop); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	defaultKubeConfig := ""
	if home := homeDir(); home != "" {
		defaultKubeConfig = filepath.Join(home, ".kube", "config")
	}
	flag.StringVar(&kubeconfig, "kubeconfig", defaultKubeConfig, "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
