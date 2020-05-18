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

// ServiceController is a controller monitoring NodePort services with
// multihome ingress enabled.
type ServiceController struct {
	factory informers.SharedInformerFactory
	lister  corelisters.ServiceLister
	synced  cache.InformerSynced
}

// NewServiceController initialises a controller for monitoring NodePort services
// with multihome ingress enabled.
func NewServiceController(clientSet kubernetes.Interface) *ServiceController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Core().V1().Services()

	controller := &ServiceController{
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

func (c *ServiceController) list() ([]*coreV1.Service, error) {
	services, err := c.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (c *ServiceController) run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting service controller.")
	defer log.Println("Shutting down service controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	services, err := c.list()
	if err != nil {
		log.Fatalf("Error listing services initially: %v", err)
	}

	addMatchingServices(services)

	<-stopChan
}

func (c *ServiceController) add(obj interface{}) {
	serviceState, ok := obj.(*coreV1.Service)
	if !ok {
		log.Printf("Could not process add: unexpected type for Service: %v", obj)
		return
	}

	addMatchingServices([]*coreV1.Service{serviceState})
}

func (c *ServiceController) update(old, new interface{}) {
	newServiceState, ok := new.(*coreV1.Service)
	if !ok {
		log.Printf("Could not process update: unexpected new state type for Service: %v", new)
		return
	}

	addMatchingServices([]*coreV1.Service{newServiceState})
}

func (c *ServiceController) delete(obj interface{}) {
	lastServiceState, ok := obj.(*coreV1.Service)
	if !ok {
		log.Printf("Could not process delete: unexpected last state type for Service: %v", obj)
		return
	}

	removeService(lastServiceState.Name, lastServiceState.Namespace)
}
