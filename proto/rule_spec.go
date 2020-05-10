package proto

import coreV1 "k8s.io/api/core/v1"

const (
	ProtocolTCP     = "tcp"     // Will result in `-p tcp -m tcp`
	ProtocolUDP     = "udp"     // Will result in `-p tcp -m udp`
	ProtocolSCTP    = "sctp"    // Will result in no specific protocol as it can be represented as either in Layer 4
	ProtocolUnknown = "unknown" // Will result in no specified protocol
)

// RuleSpecification tells the edge node how to route traffic to the NodePort specified by a matching service.
type RuleSpecification struct {
	Protocol string `json:"protocol"`
	NodePort int32  `json:"node_port"`
}

// KubeProtocolToProtoProtocol converts Kubernetes protocol representation to multihome's
func KubeProtocolToProtoProtocol(protocol coreV1.Protocol) string {
	switch protocol {
	case coreV1.ProtocolTCP:
		return ProtocolTCP
	case coreV1.ProtocolUDP:
		return ProtocolUDP
	case coreV1.ProtocolSCTP:
		return ProtocolSCTP
	default:
		return ProtocolUnknown
	}
}
