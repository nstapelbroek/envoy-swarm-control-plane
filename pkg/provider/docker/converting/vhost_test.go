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
		},
	}

	_ = collection.AddService("some_cluster", &labels)

	assert.Equal(t, collection.Vhosts["example.com"].GetDomains()[0], "example.com")
}

func TestCombinedVhostPrimaryDomainIsFirstInDomains(t *testing.T) {
	collection := NewVhostCollection()
	fLabels := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"new.example.com", "www.example.com"},
		},
	}
	bLabels := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"new.example.com", "www.example.com", "api.example.com"},
		},
	}

	_ = collection.AddService("frontend", &fLabels)
	_ = collection.AddService("backend", &bLabels)

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
			ExtraDomains: []string{"www.example.com", "api.example.com"},
			Path:         "/api",
		},
	}

	_ = collection.AddService("frontend", &fLabels)
	_ = collection.AddService("backend", &bLabels)

	assert.Equal(t, len(collection.Vhosts["example.com"].GetDomains()), 3)
}

func TestVhostDomainShouldBeUnique(t *testing.T) {
	collection := NewVhostCollection()
	firstService := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "example.com",
			ExtraDomains: []string{"example.com"},
		},
	}
	secondService := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "myexample.com",
			ExtraDomains: []string{"example.com"},
		},
	}

	assert.Equal(t, collection.AddService("first", &firstService), nil)
	assert.Error(t, collection.AddService("second", &secondService), "domain example.com is already used in another vhost")
}

func TestVhostCollectionAddServiceIsNotCorruptedOnOnError(t *testing.T) {
	collection := NewVhostCollection()
	firstService := ServiceLabel{
		Route: ServiceRoute{
			Domain: "example.com",
		},
	}
	secondService := ServiceLabel{
		Route: ServiceRoute{
			Domain:       "myexample.com",
			ExtraDomains: []string{"example.com"},
		},
	}

	assert.Equal(t, collection.AddService("first", &firstService), nil)
	// After trying to add a service that causes a validation error, vhost collection is still valid
	assert.Error(t, collection.AddService("second", &secondService), "domain example.com is already used in another vhost")
	assert.Equal(t, len(collection.Vhosts), 1)
	assert.Equal(t, len(collection.Vhosts["example.com"].GetDomains()), 1)
	assert.Check(t, collection.Vhosts["example.com"].Validate() == nil)
}
