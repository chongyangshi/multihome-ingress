package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/icydoge/multihome-ingress/controllers/apis"
	"github.com/icydoge/multihome-ingress/controllers/nodes"
	"github.com/icydoge/multihome-ingress/controllers/services"
)

type stop struct{}

const (
	ingressMultihomeNamespace = "multihome-ingress-system"
	resyncInterval            = time.Second * 30
)

var defaultKubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")

func main() {
	var (
		config *rest.Config
		err    error
	)

	if os.Getenv("KUBECONFIG") != "" {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			log.Fatalf("Cannot read kubeconfig from environment variable: %v", err.Error())
		}
		log.Printf("Using kubeconfig from environment variable location KUBECONFIG=%s", os.Getenv("KUBE_CONFIG"))
	} else {
		log.Println("No KUBECONFIG found in environment, assumi we are in cluster, using in-cluster client config.")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Could not load in-cluster config: %v", err)
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error initialising Kubernetes Node Client based on kubeconfig: %v", err)
	}

	stopChan := make(chan struct{})
	defer close(stopChan)

	nController := nodes.NewNodeController(clientSet, resyncInterval)
	go nController.Run(stopChan)

	svcController := services.NewServiceController(clientSet, resyncInterval)
	go svcController.Run(stopChan)

	ctx := context.Background()
	err := apis.Init(ctx, stopChan)
	if err != nil {
		log.Fatalf("Error initialising internal API server: %v", err)
	}

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Run forever (until interrupted)
	select {
	case <-osSignals:
		stopChan <- stop{}
	default:
	}
}
