package snapshot

import core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

const staticHash = "edge"

// StaticHash will return a constant hash for all nodes.
type StaticHash struct{}

func (StaticHash) ID(node *core.Node) string {
	if node == nil {
		return ""
	}
	return staticHash
}
