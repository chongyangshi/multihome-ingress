package controller

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/monzo/slog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	nodeV1beta1 "k8s.io/client-go/kubernetes/typed/node/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ingressMultihomeNamespace = "ingress-multihome-system"
	defaultKubeconfig         = filepath.Join(os.Getenv("HOME"), ".kube", "config")
)

func init() {
	var config *rest.Config

	if os.Getenv("KUBE_CONFIG") != "" {
		kubeconfigFile, err := ioutil.ReadFile(os.Getenv("KUBECONFIG"))
		if err != nil {
			log.Fatalf("Cannot read kubeconfig from environment variable location KUBECONFIG=%s", os.Getenv("KUBE_CONFIG"))
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Cannot parse kubeconfig from environment variable: %v", err.Error())
		}
		log.Println("Using kubeconfig from environment variable location KUBECONFIG=%s", os.Getenv("KUBE_CONFIG"))
	} else {
		log.Println("No KUBECONFIG found in environment, assumi we are in cluster, using in-cluster client config.")
		config, err := rest.InClusterConfig()
		if err != nil {
			slog.Error(ctx, "Could not load in-cluster config: %v", err)
			return err
		}
	}

	nodeClient := nodeV1beta1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error initialising Kubernetes Node Client based on kubeconfig: %v", err)
	}

	// Retrieve initial list of nodes
	var listNodesTimeout int64 = 30
	listOptions := metav1.ListOptions{TimeoutSeconds: &listNodesTimeout}
	nodes, err := nodeClient.
	if err != nil {
		log.Fatal(err)
	}

	printPVCs(pvcs)
	fmt.Println()

	// watch future changes to PVCs
	watcher, err := clientset.CoreV1().PersistentVolumeClaims(ns).Watch(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	ch := watcher.ResultChan()

}
