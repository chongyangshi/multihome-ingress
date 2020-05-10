package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const resyncInterval = time.Second * 30

type nodeController struct {
	factory informers.SharedInformerFactory
	lister  corelisters.NodeLister
	synced  cache.InformerSynced
}

type stop struct{}

const ingressMultihomeNamespace = "ingress-multihome-system"

var defaultKubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")

func newNodeController(clientSet kubernetes.Interface) *nodeController {
	informerFactory := informers.NewFilteredSharedInformerFactory(clientSet, resyncInterval, "", nil)
	informer := informerFactory.Core().V1().Nodes()

	controller := &nodeController{
		factory: informerFactory,
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.add,
		UpdateFunc: controller.update,
		DeleteFunc: controller.delete,
	})

	controller.lister = informer.Lister()
	controller.synced = informer.Informer().HasSynced

	return controller
}

func (c *nodeController) list() ([]*coreV1.Node, error) {
	nodes, err := c.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// Run starts the node controller
func (c *nodeController) run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting node controller.")
	defer log.Println("Shutting down controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	nodes, err := c.list()
	if err != nil {
		log.Fatalf("Error listing nodes initially: %v", err)
	}

	if err = createRuleSpecifications(nodes); err != nil {
		log.Fatalf("Error setting up initial rule specifications: %v", err)
	}

	<-stopChan
}

func (c *nodeController) add(obj interface{}) {
	nodeState, ok := obj.(*coreV1.Node)
	if !ok {
		log.Printf("Could not process add: unexpected type for Node: %v", obj)
		return
	}

	// @TODO
	fmt.Println(nodeState.Status.Addresses)
}

func (c *nodeController) update(old, new interface{}) {
	oldNodeState, ok := old.(*coreV1.Node)
	if !ok {
		log.Printf("Could not process update: unexpected old state type for Node: %v", old)
		return
	}
	newNodeState, ok := new.(*coreV1.Node)
	if !ok {
		log.Printf("Could not process update: unexpected new state type for Node: %v", new)
		return
	}

	// @TODO
	fmt.Println(oldNodeState.Status.Addresses)
	fmt.Println(newNodeState.Status.Addresses)
}

func (c *nodeController) delete(obj interface{}) {
	lastNodeState, ok := obj.(*coreV1.Node)
	if !ok {
		log.Printf("Could not process delete: unexpected last state type for Node: %v", obj)
		return
	}

	// @TODO
	fmt.Println(lastNodeState.Status.Addresses)
}

func main() {
	var (
		config *rest.Config
		err    error
	)

	if os.Getenv("KUBE_CONFIG") != "" {
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

	controller := newNodeController(clientSet)
	go controller.run(stopChan)

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Run forever (until interrupted)
	select {
	case <-osSignals:
		stopChan <- stop{}
	default:
	}
}
