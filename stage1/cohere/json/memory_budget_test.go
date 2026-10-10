package json

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

// Reuse markdownblocks/support_test.go's weighted budget: 512 MiB units,
// rounded up from 125% of the isolated process-tree peak. Compiler evidence
// a52b99ec measured 2.8 GiB in the parent (7 units) and 3.6 GiB per sanitized
// child (9 units). Reserve the parent once, leaving one child in the default
// 8 GiB budget, shared across chunks, parallel shards and calibration workers.
const jsonMemoryDefaultUnits = 16
const jsonMemoryParentWeight = 7
const jsonMemorySanitizedWeight = 9

var jsonMemoryOnce sync.Once
var jsonMemory *jsonMemoryBudget
var jsonMemoryError error

func jsonSanitizedMemory() (*jsonMemoryBudget, error) {
	jsonMemoryOnce.Do(func() {
		units := jsonMemoryDefaultUnits
		if value, set := os.LookupEnv("ADAMIC_JSON_MEMORY_UNITS"); set {
			var err error
			units, err = strconv.Atoi(value)
			if err != nil || units < jsonMemoryParentWeight+jsonMemorySanitizedWeight {
				jsonMemoryError = fmt.Errorf("ADAMIC_JSON_MEMORY_UNITS=%q: need at least %d (512 MiB units, including parent reserve)", value, jsonMemoryParentWeight+jsonMemorySanitizedWeight)
				return
			}
		}
		jsonMemory = newJSONMemoryBudget(units - jsonMemoryParentWeight)
	})
	return jsonMemory, jsonMemoryError
}

type jsonMemoryBudget struct {
	mu       sync.Mutex
	changed  *sync.Cond
	capacity int
	used     int
}

func newJSONMemoryBudget(capacity int) *jsonMemoryBudget {
	budget := &jsonMemoryBudget{capacity: capacity}
	budget.changed = sync.NewCond(&budget.mu)
	return budget
}

func (budget *jsonMemoryBudget) acquire(weight int) {
	budget.mu.Lock()
	defer budget.mu.Unlock()
	for budget.used+weight > budget.capacity {
		budget.changed.Wait()
	}
	budget.used += weight
}

func (budget *jsonMemoryBudget) release(weight int) {
	budget.mu.Lock()
	budget.used -= weight
	budget.changed.Broadcast()
	budget.mu.Unlock()
}
