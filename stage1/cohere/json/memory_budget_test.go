package json

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/testguard"
)

// A weighted budget in 512 MiB units, rounded up from 125% of the isolated process-tree peak. Compiler evidence
// a52b99ec measured 2.8 GiB in the parent (7 units) and 3.6 GiB per sanitized child (9 units). The parent is reserved
// once; the rest is shared across chunks, parallel shards and calibration workers. Capacity is the unit's memory
// (testguard.MemoryLimit, 16 GiB on a runner), 8 GiB only where no limit is stated: the 8 GiB constant left one
// sanitized child per process in a 16 GiB unit (#q2qjes8).
const jsonMemoryDefaultUnits = 16
const jsonMemoryUnitBytes = 512 << 20
const jsonMemoryParentWeight = 7
const jsonMemorySanitizedWeight = 9

var jsonMemoryOnce sync.Once
var jsonMemory *jsonMemoryBudget
var jsonMemoryError error

func jsonSanitizedMemory() (*jsonMemoryBudget, error) {
	jsonMemoryOnce.Do(func() {
		units := jsonMemoryDefaultUnits
		if limit, ok := testguard.MemoryLimit(); ok {
			units = int(limit / jsonMemoryUnitBytes)
		}
		if value, set := os.LookupEnv("ADAMIC_JSON_MEMORY_UNITS"); set {
			var err error
			units, err = strconv.Atoi(value)
			if err != nil || units < jsonMemoryParentWeight+jsonMemorySanitizedWeight {
				jsonMemoryError = fmt.Errorf("ADAMIC_JSON_MEMORY_UNITS=%q: need at least %d (512 MiB units, including parent reserve)", value, jsonMemoryParentWeight+jsonMemorySanitizedWeight)
				return
			}
		}
		if units < jsonMemoryParentWeight+jsonMemorySanitizedWeight {
			jsonMemoryError = fmt.Errorf("the unit's memory is %d units of 512 MiB: need at least %d, including parent reserve", units, jsonMemoryParentWeight+jsonMemorySanitizedWeight)
			return
		}
		jsonMemory = newJSONMemoryBudget(units - jsonMemoryParentWeight)
	})
	return jsonMemory, jsonMemoryError
}

// jsonMemoryBudget admits weights in arrival order. A sync.Cond that broadcasts on release let a caller arriving
// after a release take the room before a woken waiter ran again, so under load one shard waited behind every later
// shard's sanitizer run until its 90 s deadline passed (#q2qjes8: _0760, from 2.1 s to 92.1 s). Here a caller that
// finds anyone queued queues behind them, and release hands room to the head of the queue only.
type jsonMemoryBudget struct {
	mu       sync.Mutex
	capacity int
	used     int
	waiting  []jsonMemoryWaiter
}

type jsonMemoryWaiter struct {
	weight   int
	admitted chan struct{}
}

func newJSONMemoryBudget(capacity int) *jsonMemoryBudget {
	return &jsonMemoryBudget{capacity: capacity}
}

// acquire returns once weight is admitted, and how long the caller waited for it.
func (budget *jsonMemoryBudget) acquire(weight int) time.Duration {
	began := time.Now()
	budget.mu.Lock()
	if len(budget.waiting) == 0 && budget.used+weight <= budget.capacity {
		budget.used += weight
		budget.mu.Unlock()
		return time.Since(began)
	}
	waiter := jsonMemoryWaiter{weight, make(chan struct{})}
	budget.waiting = append(budget.waiting, waiter)
	budget.mu.Unlock()
	<-waiter.admitted
	return time.Since(began)
}

func (budget *jsonMemoryBudget) release(weight int) {
	budget.mu.Lock()
	defer budget.mu.Unlock()
	budget.used -= weight
	for len(budget.waiting) > 0 && budget.used+budget.waiting[0].weight <= budget.capacity {
		budget.used += budget.waiting[0].weight
		close(budget.waiting[0].admitted)
		budget.waiting = budget.waiting[1:]
	}
}

// A shard waiting for the budget is admitted before any caller that arrives after it, even one that arrives the
// instant room frees: the broadcast it replaces let that caller in first, and under load the waiter lost every time.
// Capacity admits as many weights at once as fit.
func TestJSONMemoryBudgetAdmitsInArrivalOrder(t *testing.T) {
	t.Parallel()
	budget := newJSONMemoryBudget(jsonMemorySanitizedWeight)
	budget.acquire(jsonMemorySanitizedWeight)
	admitted := make(chan string, 2)
	go func() {
		budget.acquire(jsonMemorySanitizedWeight)
		admitted <- "waiting"
		budget.release(jsonMemorySanitizedWeight)
	}()
	for queued := false; !queued; {
		budget.mu.Lock()
		queued = len(budget.waiting) == 1
		budget.mu.Unlock()
		if !queued {
			time.Sleep(time.Millisecond)
		}
	}
	budget.release(jsonMemorySanitizedWeight)
	budget.acquire(jsonMemorySanitizedWeight)
	admitted <- "arriving"
	budget.release(jsonMemorySanitizedWeight)
	if first := <-admitted; first != "waiting" {
		t.Errorf("admitted %q first, want the caller that was waiting", first)
	}

	two := newJSONMemoryBudget(2 * jsonMemorySanitizedWeight)
	for range 2 {
		if wait := two.acquire(jsonMemorySanitizedWeight); wait > time.Second {
			t.Errorf("capacity for two: an acquire waited %s with room free", wait)
		}
	}
	if len(two.waiting) != 0 || two.used != 2*jsonMemorySanitizedWeight {
		t.Errorf("capacity for two: used %d with %d waiting, want %d with none", two.used, len(two.waiting), 2*jsonMemorySanitizedWeight)
	}
}
