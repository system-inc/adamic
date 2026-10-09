package fixtures

import "testing"

// Not parallel: testFixtureDirectory owns the single Parallel call and schedules the fixture leaves.
func TestFixturesAssignmentProofs(t *testing.T) { testFixtureDirectory(t, "assignment-proofs") }

// Not parallel: testFixtureDirectory owns the single Parallel call and schedules the fixture leaves.
func TestFixturesSelfCompare(t *testing.T) { testFixtureDirectory(t, "self-compare") }
