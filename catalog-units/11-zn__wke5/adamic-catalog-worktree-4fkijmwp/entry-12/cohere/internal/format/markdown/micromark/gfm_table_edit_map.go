package micromark

import "slices"

// gfmTableEditMap is micromark-extension-gfm-table/lib/edit-map.js, upstream's port of `edit_map.rs` from
// `markdown-rs`: several changes to a list of events, batched together and applied at once.
//
// Deal with several changes in events, batching them together.
//
// Preferably, changes should be kept to a minimum.
// Sometimes, it’s needed to change the list of events, because parsing can be
// messy, and it helps to expose a cleaner interface of events to the compiler
// and other users.
// It can also help to merge many adjacent similar events.
// And, in other cases, it’s needed to parse subcontent: pass some events
// through another tokenizer and inject the result.
type gfmTableEditMap struct {
	// Record of changes.
	changes []gfmTableEditMapChange
}

// gfmTableEditMapChange is upstream's Change: at an index, remove some events and add others.
type gfmTableEditMapChange struct {
	at     int
	remove int
	add    []Event
}

// add creates an edit: a remove and/or add at a certain place. Upstream's addImplementation.
func (editMap *gfmTableEditMap) add(at int, remove int, add []Event) {
	if remove == 0 && len(add) == 0 {
		return
	}

	for index := range editMap.changes {
		if editMap.changes[index].at == at {
			editMap.changes[index].remove += remove

			// To do: before not used by tables, use when moving to micromark.
			editMap.changes[index].add = append(editMap.changes[index].add, add...)

			return
		}
	}

	editMap.changes = append(editMap.changes, gfmTableEditMapChange{at: at, remove: remove, add: add})
}

// consume is done, and changes the events. Upstream mutates the array in place; this returns the new list,
// which the caller stores where the array lived.
func (editMap *gfmTableEditMap) consume(events []Event) []Event {
	// Upstream's sort is stable; `add` merges changes at one index, so the keys are unique anyway.
	slices.SortStableFunc(editMap.changes, func(a gfmTableEditMapChange, b gfmTableEditMapChange) int {
		return a.at - b.at
	})

	if len(editMap.changes) == 0 {
		return events
	}

	// To do: if links are added in events, like they are in `markdown-rs`,
	// this is needed (upstream keeps a commented-out jump calculation here).

	index := len(editMap.changes)
	var vecs [][]Event
	for index > 0 {
		index--
		change := editMap.changes[index]
		// `events.slice(at + remove)`: past the end is an empty slice.
		from := min(change.at+change.remove, len(events))
		vecs = append(vecs, events[from:], change.add)

		// Truncate rest. A change is never past the end: the furthest is the table end, one past the
		// last event.
		events = events[:min(change.at, len(events))]
	}
	vecs = append(vecs, events)

	var result []Event
	for slice := len(vecs) - 1; slice >= 0; slice-- {
		result = append(result, vecs[slice]...)
	}

	// Truncate everything.
	editMap.changes = editMap.changes[:0]

	return result
}
