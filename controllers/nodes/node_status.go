package nodes

import (
	"sync"

	coreV1 "k8s.io/api/core/v1"
)

var (
	// Indexed by node name, which should be unique in a cluster
	readyNodesCache = map[string]*coreV1.Node{}
	readyNodesMutex = sync.RWMutex{}
)

func isNodeReady(node *coreV1.Node) bool {
	if node == nil {
		return false
	}

	for _, condition := range node.Status.Conditions {
		if condition.Type != coreV1.NodeReady {
			continue
		}

		if condition.Status != coreV1.ConditionTrue {
			return false
		}

		return true
	}

	return false
}

func updateNodeStatus(node *coreV1.Node) {
	readyNodesMutex.Lock()
	defer readyNodesMutex.Unlock()

	switch isNodeReady(node) {
	case true:
		if _, nodeAlreadyInCache := readyNodesCache[node.Name]; !nodeAlreadyInCache {
			readyNodesCache[node.Name] = node
		}
	case false:
		if _, nodeCurrentlyInCache := readyNodesCache[node.Name]; nodeCurrentlyInCache {
			delete(readyNodesCache, node.Name)
		}
	}
}

func updateNodesStatus(nodes []*coreV1.Node) {
	for _, node := range nodes {
		updateNodeStatus(node)
	}
}

func removeNode(nodeName string) {
	readyNodesMutex.Lock()
	defer readyNodesMutex.Unlock()

	delete(readyNodesCache, nodeName)
}

func getReadyNodes() []*coreV1.Node {
	readyNodesMutex.RLock()
	defer readyNodesMutex.RUnlock()

	// Due to locking, the cached nodes should be ready from the most
	// recent observation.
	var result []*coreV1.Node
	for _, node := range readyNodesCache {
		result = append(result, node)
	}

	return result
}
