package tsprinter

import (
	"sync"
	"testing"
)

type statementSharedProducts struct {
	Cases, Want, Specs, Port, CompilerHash, LoweredDir string
	Binary, Release, Library, LibraryHash              string
	Elapsed                                            float64
}

var statementCaseContexts sync.Map

// Each selected shard fetches its own inputs. No top-level setup test or
// TestMain subprocess is required, and only case work starts its deadline.
func statementReadyShared(t *testing.T) statementSharedProducts {
	t.Helper()
	return statementPrepareCommon(t)
}
