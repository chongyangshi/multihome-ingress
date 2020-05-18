package main

import (
	coreV1 "k8s.io/api/core/v1"

	"github.com/icydoge/multihome-ingress/proto"
)

func computeNodeGroup() []*proto.NodeGroup {
	// We index with a map for convenience
	var groups = map[string]*proto.NodeGroup{}

	readyNodes := getReadyNodes()
	for _, node := range readyNodes {
		groupID, ok := node.GetLabels()[proto.MultihomeIngressNodeGroupLabel]
		if !ok {
			// A node must belong to a group before we can configure its NodePorts
			// exposed to be routable.
			continue
		}

		if _, exists := groups[groupID]; !exists {
			groups[groupID] = &proto.NodeGroup{
				GroupID: groupID,
				Members: []*proto.Node{},
			}
		}

		groups[groupID].Members = append(groups[groupID].Members, &proto.Node{
			Name:        node.GetName(),
			UniqueID:    string(node.GetUID()),
			InternalIPs: getNodeInternalIPs(node.Status.Addresses),
		})
	}

	// Now serialise the map to produce a list of node groups
	var results []*proto.NodeGroup
	for _, group := range groups {
		results = append(results, group)
	}

	return results
}

// Returns all internal IPv4 addresses for a node.
func getNodeInternalIPs(addresses []coreV1.NodeAddress) []string {
	var internalIPs []string

	for _, address := range addresses {
		if address.Type != coreV1.NodeInternalIP {
			continue
		}

		internalIPs = append(internalIPs, address.Address)
	}

	return internalIPs
}
