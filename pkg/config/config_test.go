package config

import "testing"

func TestConfig(t *testing.T) {
	c := &Config{
		NodeID:       "node0",
		RaftBindPort: 9000,
	}

	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}
