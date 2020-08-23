package converting

import (
	"testing"

	"gotest.tools/assert"
)

func TestVhostPrimaryDomainIsFirstInDomains(t *testing.T) {
	collection := NewVhostCollection()
	labels := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"new.example.com", "www.example.com"},
			Path:         "/",
		},
	}

	_ = collection.AddRoute("some_cluster", &labels)

	assert.Equal(t, collection.Vhosts["example.com"].GetDomains()[0], "example.com")
}

func TestCombinedVhostPrimaryDomainIsFirstInDomains(t *testing.T) {
	collection := NewVhostCollection()
	fLabels := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"new.example.com", "www.example.com"},
			Path:         "/",
		},
	}
	bLabels := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"new.example.com", "www.example.com", "api.example.com"},
			Path:         "/api",
		},
	}

	_ = collection.AddRoute("frontend", &fLabels)
	_ = collection.AddRoute("backend", &bLabels)

	assert.Equal(t, collection.Vhosts["example.com"].GetDomains()[0], "example.com")
}

func TestCombinedVhostDomainsIsTheSumOfAllLabels(t *testing.T) {
	collection := NewVhostCollection()
	fLabels := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"www.example.com"},
			Path:         "/",
		},
	}
	bLabels := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"api.example.com"},
			Path:         "/api",
		},
	}

	_ = collection.AddRoute("frontend", &fLabels)
	_ = collection.AddRoute("backend", &bLabels)

	assert.Equal(t, len(collection.Vhosts["example.com"].GetDomains()), 3)
}
