package services

import (
	"fmt"
	"sync"

	coreV1 "k8s.io/api/core/v1"

	"github.com/icydoge/multihome-ingress/proto"
)

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

// From a list of services provided, record those that will be processed
// by multihome-ingress (NodePort services with a specific label)
func addMatchingServices(svcs []*coreV1.Service) {
	servicesMutex.Lock()
	defer servicesMutex.Unlock()

	for _, svc := range svcs {
		// We only process services with a NodePort i.e. intentionally exposed
		if svc.Spec.Type != coreV1.ServiceTypeNodePort {
			continue
		}

		// And we only process services with our specific label applied and enabled.
		if svcEnabled, found := svc.Labels[proto.MultihomeIngressServiceLabel]; !found || svcEnabled != "true" {
			continue
		}

		addOrUpdateService(svc)
	}
}

// Returns a list of services which will be processed by multihome-ingress
func getMatchingServices() []*coreV1.Service {
	servicesMutex.RLock()
	defer servicesMutex.RUnlock()

	var result []*coreV1.Service
	for _, svc := range services {
		result = append(result, svc)
	}

	return result
}

// ComputeRuleSpecifications returns a list of rule specifications to be applied on
// edges in order to route ingress for across all matching services.
func ComputeRuleSpecifications() []*proto.RuleSpecification {
	servicesMutex.RLock()
	defer servicesMutex.RUnlock()

	var result []*proto.RuleSpecification
	for _, svc := range services {
		// Should never happen as we only add NodePorts to the list
		if svc.Spec.Type != coreV1.ServiceTypeNodePort {
			continue
		}

		for _, port := range svc.Spec.Ports {
			result = append(result, &proto.RuleSpecification{
				Protocol: proto.KubeProtocolToProtoProtocol(port.Protocol),
				NodePort: port.NodePort,
			})
		}
	}

	return result
}
