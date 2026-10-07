package fresh

import (
	"maps"
	"math/rand"
	"reflect"
	"testing"
)

func TestFactsSnapshots(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(731))
	var actual facts[int, int]
	expected := map[int]int{}
	for step := 0; step < 3000; step++ {
		snapshot, previous := actual, maps.Clone(expected)
		for update := 0; update < 7; update++ {
			key := random.Intn(513) - 256
			if random.Intn(3) == 0 {
				actual.remove(key)
				delete(expected, key)
			} else {
				held := random.Int()
				actual.set(key, held)
				expected[key] = held
			}
		}
		collect := func(tree facts[int, int]) map[int]int {
			made := map[int]int{}
			for key, held := range tree.all() {
				made[key] = held
			}
			return made
		}
		if !reflect.DeepEqual(collect(snapshot), previous) {
			t.Fatal("snapshot changed")
		}
		if actual.size() != len(expected) || !reflect.DeepEqual(collect(actual), expected) {
			t.Fatalf("step %d: tree differs from map", step)
		}
		changed := map[int]int{}
		for key, held := range actual.changed(snapshot) {
			changed[key] = held
		}
		for key, held := range expected {
			if earlier, exists := previous[key]; (!exists || earlier != held) && changed[key] != held {
				t.Fatalf("step %d: changed omitted %d", step, key)
			}
		}
	}
}

// fullCloseEscapes and fullJoin are the previous whole-state union and escape
// propagation, kept here as an independent reference for snapshot operations.
func fullCloseEscapes(s *state) {
	exposed := s.escaped.sorted()
	for o := range s.heap.all() {
		if o < 0 && !s.escaped.get(o) {
			exposed = append(exposed, o)
		}
	}
	for _, o := range exposed {
		for _, held := range s.heap.get(o) {
			if s.leaked.get(o) {
				s.escape(held)
			} else {
				s.expose(held)
			}
		}
	}
}
func fullJoin(s, other *state) bool {
	changed := false
	for local, held := range other.locals.all() {
		mine := s.locals.get(local).copy()
		changed = mine.merge(held) || changed
		s.locals.set(local, mine)
	}
	for o, fields := range other.heap.all() {
		for field, held := range fields {
			changed = s.store(o, field, held) || changed
		}
	}
	for o := range other.escaped.all() {
		changed = s.escaped.add(o) || changed
	}
	for o := range other.leaked.all() {
		changed = s.leaked.add(o) || changed
	}
	if other.clobbered && !s.clobbered {
		s.clobbered, changed = true, true
	}
	if changed {
		fullCloseEscapes(s)
	}
	return changed
}

type stateObservation struct {
	locals          map[int]value
	heap            map[object]map[string]value
	escaped, leaked map[object]bool
	clobbered       bool
}

func observeState(s *state) stateObservation {
	made := stateObservation{locals: map[int]value{}, heap: map[object]map[string]value{}, escaped: map[object]bool{}, leaked: map[object]bool{}, clobbered: s.clobbered}
	for local, held := range s.locals.all() {
		made.locals[local] = held.copy()
	}
	for o, fields := range s.heap.all() {
		made.heap[o] = map[string]value{}
		for field, held := range fields {
			made.heap[o][field] = held.copy()
		}
	}
	for o := range s.escaped.all() {
		made.escaped[o] = true
	}
	for o := range s.leaked.all() {
		made.leaked[o] = true
	}
	return made
}
func TestStateSnapshotsAndEscapeDelta(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(912))
	states := []*state{newState()}
	held := func() value {
		made := value{}
		for count := random.Intn(4); count > 0; count-- {
			o := object(random.Intn(35) - 3)
			if random.Intn(3) == 0 {
				made.addWeak(o)
			} else {
				made.addStrong(o)
			}
		}
		return made
	}
	for step := 0; step < 1500; step++ {
		parent := states[random.Intn(len(states))]
		original := observeState(parent)
		current := parent.copy()
		for update := 0; update < 4; update++ {
			switch random.Intn(5) {
			case 0:
				current.locals.set(random.Intn(16), held())
			case 1:
				current.storeInto(object(random.Intn(35)-3), []string{"x", "y", elementKey}[random.Intn(3)], held())
			case 2:
				current.escape(held())
			case 3:
				current.expose(held())
			case 4:
				current.clobbered = true
			}
		}
		reference := current.copy()
		incoming := states[random.Intn(len(states))]
		changed := current.join(incoming)
		expected := fullJoin(reference, incoming)
		if changed != expected || !reflect.DeepEqual(observeState(current), observeState(reference)) {
			t.Fatalf("join differs at step %d", step)
		}
		site := random.Intn(16)
		current.recent(site)
		reference.recent(site)
		fullCloseEscapes(reference)
		if !reflect.DeepEqual(observeState(current), observeState(reference)) {
			t.Fatalf("recency differs at step %d", step)
		}
		if !reflect.DeepEqual(observeState(parent), original) {
			t.Fatalf("snapshot mutated at step %d", step)
		}
		states = append(states, current)
	}
}
