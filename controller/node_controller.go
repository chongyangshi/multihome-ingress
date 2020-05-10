package main

import (
	"log"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type nodeController struct {
	factory informers.SharedInformerFactory
	lister  corelisters.NodeLister
	synced  cache.InformerSynced
}

func newNodeController(clientSet kubernetes.Interface) *nodeController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
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

func (c *nodeController) run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting node controller.")
	defer log.Println("Shutting down node controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	nodes, err := c.list()
	if err != nil {
		log.Fatalf("Error listing nodes initially: %v", err)
	}

	updateNodesStatus(nodes)

	<-stopChan
}

func (c *nodeController) add(obj interface{}) {
	nodeState, ok := obj.(*coreV1.Node)
	if !ok {
		log.Printf("Could not process add: unexpected type for Node: %v", obj)
		return
	}

	updateNodeStatus(nodeState)

	// @TODO: retrieve new list of ready nodes and match them to DNS rules
}

func (c *nodeController) update(old, new interface{}) {
	newNodeState, ok := new.(*coreV1.Node)
	if !ok {
		log.Printf("Could not process update: unexpected new state type for Node: %v", new)
		return
	}

	updateNodeStatus(newNodeState)

	// @TODO: retrieve new list of ready nodes and match them to DNS rules
}

func (c *nodeController) delete(obj interface{}) {
	lastNodeState, ok := obj.(*coreV1.Node)
	if !ok {
		log.Printf("Could not process delete: unexpected last state type for Node: %v", obj)
		return
	}

	removeNode(lastNodeState.Name)

	// @TODO: retrieve new list of ready nodes and match them to DNS rules
}
