package proto

// NodeGroup are identified by a common group label value across all member
// nodes, whose node ports will have been routable from a common public network
// interface. This is generally the case if they run on the same underlying
// hypervisor system, and the underlying system will be responsible for routing
// ingress traffic to them.
type NodeGroup struct {
	GroupID string  `json:"group_id"`
	Members []*Node `json:"members"`
}

type Node struct {
	Name        string   `json:"name"`
	UniqueID    string   `json:"unique_id"`
	InternalIPs []string `json:"internal_ips"`
}
