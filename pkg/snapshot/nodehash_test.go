package snapshot

import (
	"testing"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"gotest.tools/assert"
)

func TestStaticHashIDReturnsEmptyStringOnNoNode(t *testing.T) {
	h := StaticHash{}

	assert.Equal(t, h.ID(nil), "")
}

func TestStaticHashIDReturnsSameHashForAnyNode(t *testing.T) {
	h := StaticHash{}

	node1 := core.Node{Id: "some-node"}
	node2 := core.Node{Id: "another-node"}

	assert.Equal(t, h.ID(&node1), h.ID(&node2))
}

func TestStaticHashIDUsesHashConstant(t *testing.T) {
	h := StaticHash{}

	assert.Equal(t, h.ID(&core.Node{Id: "1337"}), staticHash)
}
