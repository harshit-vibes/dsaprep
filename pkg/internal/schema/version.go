// Package schema provides versioned data structures for cf
package schema

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
}

var (
	// CurrentVersion is the latest schema version
	CurrentVersion = Version{Major: 1, Minor: 0, Patch: 0}

	// MinSupportedVersion is the oldest version we can migrate from
	MinSupportedVersion = Version{Major: 1, Minor: 0, Patch: 0}
)

// String returns the version as a string (e.g., "1.0.0")
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// ParseVersion parses a version string (e.g., "1.0.0")
func ParseVersion(s string) (Version, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// IsCompatible checks if two versions are compatible (same major version)
func (v Version) IsCompatible(other Version) bool {
	return v.Major == other.Major
}

// NeedsMigration checks if migration is needed between versions
func (v Version) NeedsMigration(other Version) bool {
	if v.Major != other.Major {
		return true
	}
	if v.Minor != other.Minor {
		return true
	}
	return false
}

// Compare returns -1 if v < other, 0 if v == other, 1 if v > other
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// SchemaHeader is embedded in all versioned YAML files
type SchemaHeader struct {
	Version string `yaml:"version" json:"version"`
	Type    string `yaml:"type" json:"type"`
}

// NewSchemaHeader creates a new schema header with current version
func NewSchemaHeader(schemaType string) SchemaHeader {
	return SchemaHeader{
		Version: CurrentVersion.String(),
		Type:    schemaType,
	}
}

// Schema types
const (
	TypeWorkspace  = "workspace"
	TypeProblem    = "problem"
	TypeSubmission = "submission"
	TypeProgress   = "progress"
	TypeConfig     = "config"
)
