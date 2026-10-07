package main

import (
	"fmt"
	"os"
)

const archiveUnit = "github.com/system-inc/adamic/internal/native::TestSplitTSGoAgrees"
const archiveVariable = "ADAMIC_CLANG_TSGO_ARCHIVE"

type archiveRequirement struct {
	Shard    int
	Unit     string
	Variable string
}

// Place the sole archive consumer after packing other work, including parent residuals.
func assignArchive(p *plan, weights map[string]float64) error {
	without := *p
	without.Units = nil
	index := -1
	for i, u := range p.Units {
		if u.key() == archiveUnit {
			if index >= 0 {
				return fmt.Errorf("duplicate archive consumer %s", archiveUnit)
			}
			index = i
		} else {
			without.Units = append(without.Units, u)
		}
	}
	p.Archive = nil
	if index < 0 {
		return nil
	}
	loads := predictions(without, weights)
	best := 0
	for i := 1; i < len(loads); i++ {
		if loads[i].Seconds < loads[best].Seconds {
			best = i
		}
	}
	p.Units[index].Shard = best
	p.Archive = &archiveRequirement{Shard: best, Unit: archiveUnit, Variable: archiveVariable}
	return nil
}

// Inspect coverage units as well as the declaration so a misplaced consumer cannot skip preflight.
func archiveReady(p plan, index int) error {
	for _, u := range p.Units {
		if u.key() != archiveUnit || u.Shard != index {
			continue
		}
		if p.Archive == nil || p.Archive.Shard != index || p.Archive.Unit != archiveUnit || p.Archive.Variable != archiveVariable {
			return fmt.Errorf("shard %d %s refuses: archive requirement differs from planned owner", index, archiveUnit)
		}
		path := os.Getenv(archiveVariable)
		if path == "" {
			return fmt.Errorf("shard %d %s refuses: %s missing; run archive setup separately after non-archive input setup; see docs/gate-shards.md", index, archiveUnit, archiveVariable)
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("shard %d %s refuses: %s is not a regular archive: %q", index, archiveUnit, archiveVariable, path)
		}
	}
	return nil
}
