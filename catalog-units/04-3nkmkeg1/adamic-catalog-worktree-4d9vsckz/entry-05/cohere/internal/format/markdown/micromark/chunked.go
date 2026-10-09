package micromark

import "slices"

// splice is upstream's micromark-util-chunked splice, Array#splice without the stack limit: remove
// `remove` items at `start` and insert `items` there. Returns the list, since Go slices can move.
func splice[T any](list []T, start int, remove int, items []T) []T {
	end := len(list)

	// Make start between zero and `end` (included).
	if start < 0 {
		if -start > end {
			start = 0
		} else {
			start = end + start
		}
	} else if start > end {
		start = end
	}

	if remove < 0 {
		remove = 0
	}
	if start+remove > end {
		remove = end - start
	}

	return slices.Insert(slices.Delete(list, start, start+remove), start, items...)
}

func spliceEvents(list []Event, start int, remove int, items []Event) []Event {
	return splice(list, start, remove, items)
}

// pushChunks is upstream's push: append items to list, or return items when list is empty.
func pushChunks(list []Chunk, items []Chunk) []Chunk {
	if len(list) > 0 {
		return append(list, items...)
	}
	return items
}

// pushEvents is upstream's push for events.
func pushEvents(list []Event, items []Event) []Event {
	if len(list) > 0 {
		return append(list, items...)
	}
	return items
}

// resolveAllEntries is upstream's micromark-util-resolve-all, for the tokenizer's own list.
func resolveAllEntries(entries []resolvable, events []Event, context *TokenizeContext) []Event {
	var called []*Resolver

	for _, entry := range entries {
		resolve := entry.resolveAll()

		if resolve != nil && !slices.Contains(called, resolve) {
			events = resolve.Resolve(events, context)
			called = append(called, resolve)
		}
	}

	return events
}

// resolveAll is upstream's micromark-util-resolve-all over constructs.
func resolveAll(constructs []*Construct, events []Event, context *TokenizeContext) []Event {
	var called []*Resolver

	for _, construct := range constructs {
		resolve := construct.ResolveAll

		if resolve != nil && !slices.Contains(called, resolve) {
			events = resolve.Resolve(events, context)
			called = append(called, resolve)
		}
	}

	return events
}
