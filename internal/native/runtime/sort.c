// sort.c: V8's Array.prototype.sort, TimSort, step for step.
//
// With a consistent comparator any stable sort gives the same order. With one that isn't (it reads
// state the program changes, or gives NaN, which counts as equal), the order is whatever the
// algorithm does, and so is the number and order of the comparator's calls. 0.1 means what Node
// does, so this is V8's algorithm (third_party/v8/builtins/array-sort.tq, itself CPython's
// listsort): runs found and extended by binary insertion to a minimum length, kept on a stack whose
// lengths stay balanced, and merged with galloping. The comparisons it makes, in order, are V8's;
// only the order's sign is read, as V8 reads it.

#include "adamic.h"

#ifndef ADAMIC_TARGET_WASI
#include <setjmp.h>
#endif
#include <stddef.h>
#include <stdlib.h>
#include <string.h>

// Galloping starts after this many wins in a row from one run, and the threshold adapts from here.
enum { minimum_gallop = 7 };

typedef struct {
	adamic_value *work;
	adamic_value *temporary;
	int (*compare)(adamic_value, adamic_value, void *);
	void *context;
	ptrdiff_t minimum_gallop;
	// The runs waiting to be merged: where each starts and how long it is. Their lengths grow at least
	// as fast as the Fibonacci numbers, so 85 holds any array a size_t can count.
	ptrdiff_t base[85];
	ptrdiff_t length[85];
	ptrdiff_t runs;
	// stop is where a comparator that throws sends the sort, out of however deep in a merge it is:
	// V8 sorts a copy and writes it back only once it's done, so a throw leaves the array as it was,
	// and the work in progress, perhaps halfway through a merge, is dropped.
#ifndef ADAMIC_TARGET_WASI
	jmp_buf stop;
#endif
} sort_state;

static int order(sort_state *state, adamic_value left, adamic_value right) {
	int ordered = state->compare(left, right, state->context);
	if (adamic_thrown != NULL) {
		// Only the sort's own frames lie between here and stop, none holding a reference.
#ifdef ADAMIC_TARGET_WASI
		return 0;
#else
		longjmp(state->stop, 1);
#endif
	}
	return ordered;
}

// WASI has no longjmp without experimental exception handling. Propagate the pending
// error through each sort frame before another comparison or index uses its result.
#ifdef ADAMIC_TARGET_WASI
#define STOP_SORT(value) do { if (adamic_thrown != NULL) { return value; } } while (0)
#else
#define STOP_SORT(value) ((void)0)
#endif

// copy moves count values, the right way round when the source and destination overlap.
static void copy(adamic_value *source, ptrdiff_t from, adamic_value *destination, ptrdiff_t to, ptrdiff_t count) {
	memmove(destination + to, source + from, (size_t)count * sizeof *source);
}

// minimum_run is the shortest run worth merging: n itself below 64, and otherwise between 32 and 64,
// chosen so n divided by it is a power of two or just under one.
static ptrdiff_t minimum_run(ptrdiff_t n) {
	ptrdiff_t remainder = 0;
	while (n >= 64) {
		remainder |= n & 1;
		n >>= 1;
	}
	return n + remainder;
}

// binary_insertion sorts [low, high), of which [low, start) is already sorted.
static void binary_insertion(sort_state *state, ptrdiff_t low, ptrdiff_t start, ptrdiff_t high) {
	adamic_value *work = state->work;
	if (low == start) {
		start++;
	}
	for (; start < high; start++) {
		ptrdiff_t left = low, right = start;
		adamic_value pivot = work[start];
		while (left < right) {
			ptrdiff_t middle = left + ((right - left) >> 1);
			int ordered = order(state, pivot, work[middle]);
			STOP_SORT();
			if (ordered < 0) {
				right = middle;
			} else {
				left = middle + 1;
			}
		}
		for (ptrdiff_t place = start; place > left; place--) {
			work[place] = work[place - 1];
		}
		work[left] = pivot;
	}
}

// count_run is the length of the run starting at low, made ascending: a strictly descending run is
// reversed, which keeps the sort stable since no two of its elements are equal.
static ptrdiff_t count_run(sort_state *state, ptrdiff_t low, ptrdiff_t high) {
	adamic_value *work = state->work;
	ptrdiff_t next = low + 1;
	if (next == high) {
		return 1;
	}
	ptrdiff_t length = 2;
	adamic_value previous = work[next];
	int ordered = order(state, work[next], work[next - 1]);
	STOP_SORT(0);
	bool descending = ordered < 0;
	for (ptrdiff_t index = next + 1; index < high; index++) {
		adamic_value current = work[index];
		int result = order(state, current, previous);
		STOP_SORT(0);
		if (descending ? result >= 0 : result < 0) {
			break;
		}
		previous = current;
		length++;
	}
	if (descending) {
		for (ptrdiff_t left = low, right = low + length - 1; left < right; left++, right--) {
			adamic_value swapped = work[left];
			work[left] = work[right];
			work[right] = swapped;
		}
	}
	return length;
}

// gallop_left finds where key goes in the sorted [base, base + length): the leftmost place, before
// any element equal to it. It starts looking at base + hint.
static ptrdiff_t gallop_left(sort_state *state, adamic_value *array, adamic_value key, ptrdiff_t base, ptrdiff_t length, ptrdiff_t hint) {
	ptrdiff_t last = 0, offset = 1;
	int ordered = order(state, array[base + hint], key);
	STOP_SORT(0);
	if (ordered < 0) {
		ptrdiff_t most = length - hint;
		while (offset < most) {
			int ordered = order(state, array[base + hint + offset], key);
			STOP_SORT(0);
			if (ordered >= 0) {
				break;
			}
			last = offset;
			offset = (offset << 1) + 1;
			if (offset <= 0) {
				offset = most;
			}
		}
		if (offset > most) {
			offset = most;
		}
		last += hint;
		offset += hint;
	} else {
		ptrdiff_t most = hint + 1;
		while (offset < most) {
			int ordered = order(state, array[base + hint - offset], key);
			STOP_SORT(0);
			if (ordered < 0) {
				break;
			}
			last = offset;
			offset = (offset << 1) + 1;
			if (offset <= 0) {
				offset = most;
			}
		}
		if (offset > most) {
			offset = most;
		}
		ptrdiff_t swapped = last;
		last = hint - offset;
		offset = hint - swapped;
	}
	last++;
	while (last < offset) {
		ptrdiff_t middle = last + ((offset - last) >> 1);
		int ordered = order(state, array[base + middle], key);
		STOP_SORT(0);
		if (ordered < 0) {
			last = middle + 1;
		} else {
			offset = middle;
		}
	}
	return offset;
}

// gallop_right is gallop_left's twin: the rightmost place, after any element equal to key.
static ptrdiff_t gallop_right(sort_state *state, adamic_value *array, adamic_value key, ptrdiff_t base, ptrdiff_t length, ptrdiff_t hint) {
	ptrdiff_t last = 0, offset = 1;
	int ordered = order(state, key, array[base + hint]);
	STOP_SORT(0);
	if (ordered < 0) {
		ptrdiff_t most = hint + 1;
		while (offset < most) {
			int ordered = order(state, key, array[base + hint - offset]);
			STOP_SORT(0);
			if (ordered >= 0) {
				break;
			}
			last = offset;
			offset = (offset << 1) + 1;
			if (offset <= 0) {
				offset = most;
			}
		}
		if (offset > most) {
			offset = most;
		}
		ptrdiff_t swapped = last;
		last = hint - offset;
		offset = hint - swapped;
	} else {
		ptrdiff_t most = length - hint;
		while (offset < most) {
			int ordered = order(state, key, array[base + hint + offset]);
			STOP_SORT(0);
			if (ordered < 0) {
				break;
			}
			last = offset;
			offset = (offset << 1) + 1;
			if (offset <= 0) {
				offset = most;
			}
		}
		if (offset > most) {
			offset = most;
		}
		last += hint;
		offset += hint;
	}
	last++;
	while (last < offset) {
		ptrdiff_t middle = last + ((offset - last) >> 1);
		int ordered = order(state, key, array[base + middle]);
		STOP_SORT(0);
		if (ordered < 0) {
			offset = middle;
		} else {
			last = middle + 1;
		}
	}
	return offset;
}

// merge_low merges the adjacent runs A and B, A the shorter, front to back, with A copied aside.
static void merge_low(sort_state *state, ptrdiff_t base_a, ptrdiff_t length_a, ptrdiff_t base_b, ptrdiff_t length_b) {
	adamic_value *work = state->work, *temporary = state->temporary;
	copy(work, base_a, temporary, 0, length_a);
	ptrdiff_t destination = base_a, cursor_temporary = 0, cursor_b = base_b;
	work[destination++] = work[cursor_b++];
	if (--length_b == 0) {
		goto succeed;
	}
	if (length_a == 1) {
		goto copy_b;
	}
	ptrdiff_t gallop = state->minimum_gallop;
	for (;;) {
		ptrdiff_t wins_a = 0, wins_b = 0;
		for (;;) {
			int ordered = order(state, work[cursor_b], temporary[cursor_temporary]);
			STOP_SORT();
			if (ordered < 0) {
				work[destination++] = work[cursor_b++];
				wins_b++;
				length_b--;
				wins_a = 0;
				if (length_b == 0) {
					goto succeed;
				}
				if (wins_b >= gallop) {
					break;
				}
			} else {
				work[destination++] = temporary[cursor_temporary++];
				wins_a++;
				length_a--;
				wins_b = 0;
				if (length_a == 1) {
					goto copy_b;
				}
				if (wins_a >= gallop) {
					break;
				}
			}
		}
		gallop++;
		bool first = true;
		while (wins_a >= minimum_gallop || wins_b >= minimum_gallop || first) {
			first = false;
			gallop = gallop - 1 > 1 ? gallop - 1 : 1;
			state->minimum_gallop = gallop;
			wins_a = gallop_right(state, temporary, work[cursor_b], cursor_temporary, length_a, 0);
			STOP_SORT();
			if (wins_a > 0) {
				copy(temporary, cursor_temporary, work, destination, wins_a);
				destination += wins_a;
				cursor_temporary += wins_a;
				length_a -= wins_a;
				if (length_a == 1) {
					goto copy_b;
				}
				// Impossible with a consistent comparator, which this may not be.
				if (length_a == 0) {
					goto succeed;
				}
			}
			work[destination++] = work[cursor_b++];
			if (--length_b == 0) {
				goto succeed;
			}
			wins_b = gallop_left(state, work, temporary[cursor_temporary], cursor_b, length_b, 0);
			STOP_SORT();
			if (wins_b > 0) {
				copy(work, cursor_b, work, destination, wins_b);
				destination += wins_b;
				cursor_b += wins_b;
				length_b -= wins_b;
				if (length_b == 0) {
					goto succeed;
				}
			}
			work[destination++] = temporary[cursor_temporary++];
			if (--length_a == 1) {
				goto copy_b;
			}
		}
		gallop++;
		state->minimum_gallop = gallop;
	}
succeed:
	if (length_a > 0) {
		copy(temporary, cursor_temporary, work, destination, length_a);
	}
	return;
copy_b:
	// The last of A belongs after all of what's left of B.
	copy(work, cursor_b, work, destination, length_b);
	work[destination + length_b] = temporary[cursor_temporary];
}

// merge_high merges the adjacent runs A and B, B the shorter, back to front, with B copied aside.
static void merge_high(sort_state *state, ptrdiff_t base_a, ptrdiff_t length_a, ptrdiff_t base_b, ptrdiff_t length_b) {
	adamic_value *work = state->work, *temporary = state->temporary;
	copy(work, base_b, temporary, 0, length_b);
	ptrdiff_t destination = base_b + length_b - 1, cursor_temporary = length_b - 1, cursor_a = base_a + length_a - 1;
	work[destination--] = work[cursor_a--];
	if (--length_a == 0) {
		goto succeed;
	}
	if (length_b == 1) {
		goto copy_a;
	}
	ptrdiff_t gallop = state->minimum_gallop;
	for (;;) {
		ptrdiff_t wins_a = 0, wins_b = 0;
		for (;;) {
			int ordered = order(state, temporary[cursor_temporary], work[cursor_a]);
			STOP_SORT();
			if (ordered < 0) {
				work[destination--] = work[cursor_a--];
				wins_a++;
				length_a--;
				wins_b = 0;
				if (length_a == 0) {
					goto succeed;
				}
				if (wins_a >= gallop) {
					break;
				}
			} else {
				work[destination--] = temporary[cursor_temporary--];
				wins_b++;
				length_b--;
				wins_a = 0;
				if (length_b == 1) {
					goto copy_a;
				}
				if (wins_b >= gallop) {
					break;
				}
			}
		}
		gallop++;
		bool first = true;
		while (wins_a >= minimum_gallop || wins_b >= minimum_gallop || first) {
			first = false;
			gallop = gallop - 1 > 1 ? gallop - 1 : 1;
			state->minimum_gallop = gallop;
			ptrdiff_t found = gallop_right(state, work, temporary[cursor_temporary], base_a, length_a, length_a - 1);
			STOP_SORT();
			wins_a = length_a - found;
			if (wins_a > 0) {
				destination -= wins_a;
				cursor_a -= wins_a;
				copy(work, cursor_a + 1, work, destination + 1, wins_a);
				length_a -= wins_a;
				if (length_a == 0) {
					goto succeed;
				}
			}
			work[destination--] = temporary[cursor_temporary--];
			if (--length_b == 1) {
				goto copy_a;
			}
			found = gallop_left(state, temporary, work[cursor_a], 0, length_b, length_b - 1);
			STOP_SORT();
			wins_b = length_b - found;
			if (wins_b > 0) {
				destination -= wins_b;
				cursor_temporary -= wins_b;
				copy(temporary, cursor_temporary + 1, work, destination + 1, wins_b);
				length_b -= wins_b;
				if (length_b == 1) {
					goto copy_a;
				}
				// Impossible with a consistent comparator, which this may not be.
				if (length_b == 0) {
					goto succeed;
				}
			}
			work[destination--] = work[cursor_a--];
			if (--length_a == 0) {
				goto succeed;
			}
		}
		gallop++;
		state->minimum_gallop = gallop;
	}
succeed:
	if (length_b > 0) {
		copy(temporary, 0, work, destination - (length_b - 1), length_b);
	}
	return;
copy_a:
	// The first of B belongs before all of what's left of A.
	destination -= length_a;
	cursor_a -= length_a;
	copy(work, cursor_a + 1, work, destination + 1, length_a);
	work[destination] = temporary[cursor_temporary];
}

// merge_at merges the runs at i and i + 1 on the stack. What's already in place at either end is
// found by galloping and left where it is.
static void merge_at(sort_state *state, ptrdiff_t i) {
	ptrdiff_t base_a = state->base[i], length_a = state->length[i];
	ptrdiff_t base_b = state->base[i + 1], length_b = state->length[i + 1];
	state->length[i] = length_a + length_b;
	if (i == state->runs - 3) {
		state->base[i + 1] = state->base[i + 2];
		state->length[i + 1] = state->length[i + 2];
	}
	state->runs--;
	ptrdiff_t skipped = gallop_right(state, state->work, state->work[base_b], base_a, length_a, 0);
	STOP_SORT();
	base_a += skipped;
	length_a -= skipped;
	if (length_a == 0) {
		return;
	}
	length_b = gallop_left(state, state->work, state->work[base_a + length_a - 1], base_b, length_b, length_b - 1);
	STOP_SORT();
	if (length_b == 0) {
		return;
	}
	if (length_a <= length_b) {
		merge_low(state, base_a, length_a, base_b, length_b);
		STOP_SORT();
	} else {
		merge_high(state, base_a, length_a, base_b, length_b);
		STOP_SORT();
	}
}

// balanced says whether the run at n is longer than the two above it together.
static bool balanced(sort_state *state, ptrdiff_t n) {
	if (n < 2) {
		return true;
	}
	return state->length[n - 2] > state->length[n - 1] + state->length[n];
}

// merge_collapse merges until the stack's lengths are balanced again.
static void merge_collapse(sort_state *state) {
	while (state->runs > 1) {
		ptrdiff_t n = state->runs - 2;
		if (!balanced(state, n + 1) || !balanced(state, n)) {
			if (state->length[n - 1] < state->length[n + 1]) {
				n--;
			}
			merge_at(state, n);
			STOP_SORT();
		} else if (state->length[n] <= state->length[n + 1]) {
			merge_at(state, n);
			STOP_SORT();
		} else {
			break;
		}
	}
}

// merge_force_collapse merges everything left, at the end.
static void merge_force_collapse(sort_state *state) {
	while (state->runs > 1) {
		ptrdiff_t n = state->runs - 2;
		if (n > 0 && state->length[n - 1] < state->length[n + 1]) {
			n--;
		}
		merge_at(state, n);
		STOP_SORT();
	}
}

// sorted is the sort itself, which a comparator that throws leaves by stop.
static void sorted(sort_state *state, ptrdiff_t length) {
	ptrdiff_t remaining = length, low = 0;
	ptrdiff_t shortest = minimum_run(remaining);
	while (remaining != 0) {
		ptrdiff_t run = count_run(state, low, low + remaining);
		STOP_SORT();
		if (run < shortest) {
			ptrdiff_t forced = shortest < remaining ? shortest : remaining;
			binary_insertion(state, low, low + run, low + forced);
			STOP_SORT();
			run = forced;
		}
		state->base[state->runs] = low;
		state->length[state->runs] = run;
		state->runs++;
		merge_collapse(state);
		STOP_SORT();
		low += run;
		remaining -= run;
	}
	merge_force_collapse(state);
	STOP_SORT();
}

// sorted_or_stopped runs the sort, and says false when a comparator threw. The setjmp is in a frame
// of its own, whose locals nothing changes after it, so none is left indeterminate by the longjmp.
static bool sorted_or_stopped(sort_state *state, ptrdiff_t length) {
#ifndef ADAMIC_TARGET_WASI
	if (setjmp(state->stop) != 0) {
		return false;
	}
#endif
	sorted(state, length);
	STOP_SORT(false);
	return true;
}

bool adamic_timsort(adamic_value *work, size_t count, int (*compare)(adamic_value, adamic_value, void *), void *context) {
	ptrdiff_t length = (ptrdiff_t)count;
	if (length < 2) {
		return true;
	}
	sort_state state = {.work = work, .compare = compare, .context = context, .minimum_gallop = minimum_gallop};
	state.temporary = malloc(count * sizeof *work);
	if (state.temporary == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	bool finished = sorted_or_stopped(&state, length);
	free(state.temporary);
	return finished;
}
