package proto

// These labels are used to identify what nodes and what services should have
// multihome ingress support.
const (
	// MultihomeIngressServiceLabel defines whether a NodePort service should be subject
	// to multihome ingress support, with string value "true" denoting enabled.
	MultihomeIngressServiceLabel = "multihome-ingress.kube-system.com/service"

	// MultihomeIngressNodeLabel defines whether a Node should be have multihome
	// ingress support, with string value "true" denoting enabled.
	MultihomeIngressNodeLabel = "multihome-ingress.kube-system.com/node"

	// MultihomeIngressNodeGroupLabel defines what group each node supporting multihome
	// ingress belongs to. Each group of nodes share the same public IP for ingress
	// via the internet.
	MultihomeIngressNodeGroupLabel = "multihome-ingress.kube-system.com/node-group"
)
