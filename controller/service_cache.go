package main

import (
	"fmt"
	"sync"

	coreV1 "k8s.io/api/core/v1"
)

const serviceLabel = "multihome-ingress"

var (
	services      = map[string]*coreV1.Service{}
	servicesMutex = sync.RWMutex{}
)

func addOrUpdateService(svc *coreV1.Service) {
	if svc == nil {
		return
	}

	servicesMutex.Lock()
	defer servicesMutex.Unlock()

	services[fmt.Sprintf("%s:%s", svc.Namespace, svc.Name)] = svc
}

func removeService(name, namespace string) {
	servicesMutex.Lock()
	defer servicesMutex.Unlock()

	key := fmt.Sprintf("%s:%s", namespace, name)
	if _, found := services[key]; !found {
		return
	}

	delete(services, key)
}

func getService(name, namespace string) (*coreV1.Service, bool) {
	servicesMutex.RLock()
	defer servicesMutex.RUnlock()

	svc, ok := services[fmt.Sprintf("%s:%s", namespace, name)]
	return svc, ok
}

// Returns a list of services which will be processed by ingress-multihome
func getMatchingServices() []*coreV1.Service {
	servicesMutex.RLock()
	defer servicesMutex.RUnlock()

	var result []*coreV1.Service
	for _, svc := range services {
		result = append(result, svc)
	}

	return result
}

// From a list of services provided, record those that will be processed
// by k8s-ingress-multihome (NodePort services with a specific label)
func addMatchingServices(svcs []*coreV1.Service) {
	for _, svc := range svcs {
		// We only process services with a NodePort i.e. intentionally exposed
		if svc.Spec.Type != coreV1.ServiceTypeNodePort {
			continue
		}

		// And we only process services with our specific label applied
		if _, found := svc.Labels[serviceLabel]; !found {
			continue
		}

		addOrUpdateService(svc)
	}
}
