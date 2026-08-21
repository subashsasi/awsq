package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestParseEC2Filters(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		checkFirst    string // expected AWS filter name for first entry
	}{
		{"state filter", "state=running", 1, "instance-state-name"},
		{"type filter", "type=t3.micro", 1, "instance-type"},
		{"name filter", "name=web-server", 1, "tag:Name"},
		{"vpc filter", "vpc=vpc-123", 1, "vpc-id"},
		{"az filter", "az=us-east-1a", 1, "availability-zone"},
		{"custom tag filter", "Environment=prod", 1, "tag:Environment"},
		{"multiple filters", "state=running,type=t3.micro", 2, "instance-state-name"},
		{"invalid pair ignored", "noequals", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEC2Filters(tt.input)
			if len(result) != tt.expectedCount {
				t.Errorf("parseEC2Filters(%q) returned %d filters, want %d", tt.input, len(result), tt.expectedCount)
				return
			}
			if tt.expectedCount > 0 && *result[0].Name != tt.checkFirst {
				t.Errorf("parseEC2Filters(%q)[0].Name = %q, want %q", tt.input, *result[0].Name, tt.checkFirst)
			}
		})
	}
}

func TestGetTagValue(t *testing.T) {
	tags := []types.Tag{
		{Key: strPtr("Name"), Value: strPtr("web-server")},
		{Key: strPtr("Environment"), Value: strPtr("prod")},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing tag", "Name", "web-server"},
		{"another existing tag", "Environment", "prod"},
		{"missing tag", "Team", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTagValue(tags, tt.key)
			if result != tt.expected {
				t.Errorf("getTagValue(tags, %q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetPlacementAZ(t *testing.T) {
	az := "us-east-1a"
	tests := []struct {
		name      string
		placement *types.Placement
		expected  string
	}{
		{"nil placement", nil, "-"},
		{"valid placement", &types.Placement{AvailabilityZone: &az}, "us-east-1a"},
		{"placement with nil AZ", &types.Placement{AvailabilityZone: nil}, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPlacementAZ(tt.placement)
			if result != tt.expected {
				t.Errorf("getPlacementAZ() = %q, want %q", result, tt.expected)
			}
		})
	}
}
