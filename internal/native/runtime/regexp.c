// Compile-time bytecode, executed over UTF-16 with ordered backtracking.
#include "adamic.h"
#include "regexp_fold.h"
#include <math.h>
#include <stdlib.h>
#include <string.h>

static uint64_t regex_step_limit;
static int regex_regular_mode = 1;
void adamic_regex_set_regular_enabled(bool enabled) { regex_regular_mode = enabled ? 1 : 0; }
void adamic_regex_set_regular_mode(int mode) { regex_regular_mode = mode; }
void adamic_regex_set_step_limit(uint64_t limit) { regex_step_limit = limit; }
static void *regex_memory(size_t size) {
	void *memory = malloc(size == 0 ? 1 : size);
	if (memory == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return memory;
}
static bool high(uint16_t c) { return c >= 0xd800 && c <= 0xdbff; }
static bool low(uint16_t c) { return c >= 0xdc00 && c <= 0xdfff; }
static bool regex_read(const uint16_t *input, size_t length, ptrdiff_t at, int direction,
					   bool unicode, uint32_t *point, ptrdiff_t *next) {
	if (unicode && at > 0 && (size_t)at < length && high(input[at - 1]) && low(input[at]))
		return false;
	if (direction > 0) {
		if ((size_t)at >= length)
			return false;
		uint32_t c = input[at];
		*next = at + 1;
		if (unicode && high((uint16_t)c) && (size_t)(at + 1) < length && low(input[at + 1])) {
			c = 0x10000 + ((c - 0xd800) << 10) + input[at + 1] - 0xdc00;
			(*next)++;
		}
		*point = c;
		return true;
	}
	if (at <= 0)
		return false;
	uint32_t c = input[at - 1];
	*next = at - 1;
	if (unicode && low((uint16_t)c) && at > 1 && high(input[at - 2])) {
		c = 0x10000 + (((uint32_t)input[at - 2] - 0xd800) << 10) + c - 0xdc00;
		(*next)--;
	}
	*point = c;
	return true;
}
static uint32_t regex_canonical(uint32_t c, unsigned flags) {
	if (!(flags & 1))
		return c;
	if (c < 128) {
		if (flags & 4)
			return c >= 'A' && c <= 'Z' ? c + ('a' - 'A') : c;
		return c >= 'a' && c <= 'z' ? c - ('a' - 'A') : c;
	}

	const uint32_t (*table)[2] = flags & 4 ? regex_unicode_fold : regex_legacy_fold;
	size_t count = flags & 4 ? sizeof regex_unicode_fold / sizeof regex_unicode_fold[0]
							 : sizeof regex_legacy_fold / sizeof regex_legacy_fold[0];
	size_t first = 0, end = count;
	while (first < end) {
		size_t mid = first + (end - first) / 2;
		if (table[mid][0] < c)
			first = mid + 1;
		else
			end = mid;
	}
	return first < count && table[first][0] == c ? table[first][1] : c;
}
static bool regex_word(uint32_t c, unsigned flags) {
	if ((flags & 5) == 5)
		c = regex_canonical(c, flags);
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z');
}
static bool regex_line(uint16_t c) { return c == 10 || c == 13 || c == 0x2028 || c == 0x2029; }
static bool regex_contains(const adamic_regex_instruction *i, uint32_t c) {
	size_t first = 0, end = i->range_count;
	while (first < end) {
		size_t mid = first + (end - first) / 2;
		if (i->ranges[mid].last < c)
			first = mid + 1;
		else
			end = mid;
	}
	return first < i->range_count && i->ranges[first].first <= c;
}
// Keep the VM's helpers local so clang can specialize them. The regular
// engine shares their semantics through these out-of-line entry points.
bool adamic_regex_read(const uint16_t *input, size_t length, ptrdiff_t at, int direction,
					   bool unicode, uint32_t *point, ptrdiff_t *next) {
	return regex_read(input, length, at, direction, unicode, point, next);
}
uint32_t adamic_regex_canonical(uint32_t c, unsigned flags) { return regex_canonical(c, flags); }
bool adamic_regex_word(uint32_t c, unsigned flags) { return regex_word(c, flags); }
bool adamic_regex_contains(const adamic_regex_instruction *i, uint32_t c) {
	return regex_contains(i, c);
}
typedef struct {
	uint64_t count;
	ptrdiff_t start;
} regex_repeat;
typedef struct regex_workspace regex_workspace;
typedef struct regex_state {
	size_t pc;
	ptrdiff_t position;
	ptrdiff_t *captures;
	regex_repeat *repeats;
	struct regex_state *previous;
	_Alignas(regex_repeat) bool has_frame;
	bool allocation_on_heap;
	size_t register_bytes;
	regex_workspace *workspace;
} regex_state;
// Choice points own their snapshots until popped. Freed frames can be reused
// across search positions; overflow blocks are released at execution end.
struct regex_workspace {
	regex_state *free_frames;
	union {
		max_align_t alignment;
		unsigned char bytes[8192];
	} arena;
	size_t used;
};
static void regex_workspace_destroy(regex_workspace *workspace) {
	while (workspace->free_frames != NULL) {
		regex_state *frame = workspace->free_frames;
		workspace->free_frames = frame->previous;
		if (frame->allocation_on_heap)
			free(frame);
	}
}
static regex_state regex_clone(const regex_state *state, const adamic_regex_program *p) {
	regex_state result = *state;
	// One aligned block holds a choice point and both register snapshots.
	size_t bytes =
		2 * (p->captures + 1) * sizeof *result.captures + p->repeats * sizeof *result.repeats;
	regex_workspace *workspace = state->workspace;
	regex_state **available = &workspace->free_frames;
	while (*available != NULL && (*available)->register_bytes < bytes)
		available = &(*available)->previous;
	regex_state *frame = *available;
	if (frame != NULL) {
		*available = frame->previous;
		result.register_bytes = frame->register_bytes;
		result.allocation_on_heap = frame->allocation_on_heap;
	} else {
		size_t alignment = _Alignof(regex_state);
		size_t size = (sizeof *frame + bytes + alignment - 1) / alignment * alignment;
		if (size <= sizeof workspace->arena.bytes - workspace->used) {
			frame = (regex_state *)(workspace->arena.bytes + workspace->used);
			workspace->used += size;
			result.allocation_on_heap = false;
		} else {
			frame = regex_memory(size);
			result.allocation_on_heap = true;
		}
		result.register_bytes = bytes;
	}
	result.captures = (ptrdiff_t *)(frame + 1);
	result.repeats = (regex_repeat *)(result.captures + 2 * (p->captures + 1));
	memcpy(result.captures, state->captures, 2 * (p->captures + 1) * sizeof *result.captures);
	if (state->repeats != NULL)
		memcpy(result.repeats, state->repeats, p->repeats * sizeof *result.repeats);
	else
		memset(result.repeats, 0, p->repeats * sizeof *result.repeats);
	result.has_frame = true;
	result.previous = NULL;
	return result;
}
static void regex_free(regex_state state) {
	if (state.has_frame) {
		regex_state *frame = (regex_state *)state.captures - 1;
		*frame = state;
		frame->previous = state.workspace->free_frames;
		state.workspace->free_frames = frame;
	}
}
static void regex_push(regex_state **stack, regex_state state) {
	regex_state *frame = (regex_state *)state.captures - 1;
	*frame = state;
	frame->previous = *stack;
	*stack = frame;
}
static bool regex_run(const adamic_regex_program *p, const uint16_t *input, size_t length,
					  ptrdiff_t *position, ptrdiff_t *captures, uint64_t *steps,
					  regex_workspace *workspace) {
	ADAMIC_CHECK_STACK();
	ptrdiff_t local_captures[32];
	regex_repeat local_repeats[16];
	regex_state state = {.position = *position, .captures = captures, .workspace = workspace};
	if (2 * (p->captures + 1) <= 32 && p->repeats <= 16) {
		state.captures = local_captures;
		state.repeats = local_repeats;
		memcpy(state.captures, captures, 2 * (p->captures + 1) * sizeof *captures);
		memset(state.repeats, 0, p->repeats * sizeof *state.repeats);
	} else
		state = regex_clone(&state, p);
	regex_state *stack = NULL;
	bool matched = false;
	for (;;) {
		if (regex_step_limit && *steps >= regex_step_limit) {
			static const char message[] = "regexp: instruction step limit exceeded";
			adamic_panic(message, sizeof message - 1);
		}
		(*steps)++;
		const adamic_regex_instruction *i = &p->code[state.pc];
		bool failed = false;
		switch (i->op) {
		case 0:
			matched = true;
			goto finished;
		case 3:
			state.pc = (size_t)i->x;
			continue;
		case 2: {
			regex_state alternate = regex_clone(&state, p);
			alternate.pc = (size_t)i->y;
			regex_push(&stack, alternate);
			state.pc = (size_t)i->x;
			continue;
		}
		case 4:
			state.captures[i->x] = state.position;
			break;
		case 1: {
			// String alternatives are sorted by consumed length, then the code point.
			ptrdiff_t singleton;
			ptrdiff_t *positions = i->string_count == 0
									   ? &singleton
									   : regex_memory((i->string_count + 1) * sizeof *positions);
			size_t count = 0;
			for (size_t k = 0; k < i->string_count; k++) {
				ptrdiff_t at = state.position;
				bool ok = true;
				const adamic_regex_text *text = &i->strings[k];
				for (size_t j = 0; j < text->count; j++) {
					uint32_t c;
					ptrdiff_t next;
					size_t offset = i->direction < 0 ? text->count - 1 - j : j;
					if (!regex_read(input, length, at, i->direction, i->flags & 4, &c, &next) ||
						regex_canonical(c, i->flags) != text->points[offset]) {
						ok = false;
						break;
					}
					at = next;
				}
				if (ok)
					positions[count++] = at;
			}
			uint32_t c;
			ptrdiff_t next;
			if (regex_read(input, length, state.position, i->direction, i->flags & 4, &c, &next) &&
				regex_contains(i, regex_canonical(c, i->flags)))
				positions[count++] = next;
			for (size_t k = 1; k < count; k++) {
				ptrdiff_t value = positions[k];
				size_t j = k;
				while (j > 0 &&
					   llabs(value - state.position) > llabs(positions[j - 1] - state.position)) {
					positions[j] = positions[j - 1];
					j--;
				}
				positions[j] = value;
			}
			// A class is a set: equal endpoints are one alternative, including strings
			// equal after Canonicalize. Repeating duplicate choices is exponential.
			size_t unique = 0;
			for (size_t k = 0; k < count; k++)
				if (unique == 0 || positions[k] != positions[unique - 1])
					positions[unique++] = positions[k];
			count = unique;
			if (count == 0)
				failed = true;
			else {
				for (size_t k = count; k > 1; k--) {
					regex_state alt = regex_clone(&state, p);
					alt.pc++;
					alt.position = positions[k - 1];
					regex_push(&stack, alt);
				}
				state.position = positions[0];
			}
			if (i->string_count != 0)
				free(positions);
			break;
		}
		case 5:
			if (i->assertion == 0)
				failed = !(state.position == 0 || ((i->flags & 2) && state.position > 0 &&
												   regex_line(input[state.position - 1])));
			else if (i->assertion == 1)
				failed = !((size_t)state.position == length ||
						   ((i->flags & 2) && (size_t)state.position < length &&
							regex_line(input[state.position])));
			else {
				uint32_t left = 0, right = 0;
				ptrdiff_t next;
				bool l = regex_read(input, length, state.position, -1, i->flags & 4, &left, &next),
					 r = regex_read(input, length, state.position, 1, i->flags & 4, &right, &next);
				bool boundary =
					(l && regex_word(left, i->flags)) != (r && regex_word(right, i->flags));
				failed = boundary == (i->assertion == 3);
			}
			break;
		case 6: {
			ptrdiff_t subposition = state.position;
			ptrdiff_t local_sub[32];
			ptrdiff_t *sub =
				p->captures < 16 ? local_sub : regex_memory(2 * (p->captures + 1) * sizeof *sub);
			memcpy(sub, state.captures, 2 * (p->captures + 1) * sizeof *sub);
			bool ok = regex_run(i->look, input, length, &subposition, sub, steps, workspace);
			failed = ok == i->negative;
			if (ok && !i->negative)
				memcpy(state.captures, sub, 2 * (p->captures + 1) * sizeof *sub);
			if (sub != local_sub)
				free(sub);
			break;
		}
		case 7: {
			ptrdiff_t start = -1, end = -1;
			for (size_t k = 0; k < i->id_count; k++) {
				size_t id = i->ids[k];
				if (state.captures[2 * id] >= 0 && state.captures[2 * id + 1] >= 0) {
					start = state.captures[2 * id];
					end = state.captures[2 * id + 1];
				}
			}
			if (start >= 0) {
				ptrdiff_t at = i->direction < 0 ? end : start,
						  target = i->direction < 0 ? start : end;
				while (at != target) {
					uint32_t a, b;
					ptrdiff_t next, to;
					if (!regex_read(input, length, at, i->direction, i->flags & 4, &a, &next) ||
						!regex_read(input, length, state.position, i->direction, i->flags & 4, &b,
									&to) ||
						regex_canonical(a, i->flags) != regex_canonical(b, i->flags)) {
						failed = true;
						break;
					}
					at = next;
					state.position = to;
				}
			}
			break;
		}
		case 8:
			state.repeats[i->x] = (regex_repeat){0, 0};
			break;
		case 9: {
			regex_repeat reg = state.repeats[i->x];
			bool exit = reg.count >= i->minimum, body = i->unbounded || reg.count < i->maximum;
			if (!body && !exit) {
				failed = true;
				break;
			}
			if (body && exit) {
				regex_state alternate = regex_clone(&state, p);
				if (i->greedy)
					alternate.pc = (size_t)i->y;
				else {
					alternate.pc++;
					alternate.repeats[i->x].start = state.position;
					for (size_t k = 0; k < i->id_count; k++) {
						alternate.captures[2 * i->ids[k]] = -1;
						alternate.captures[2 * i->ids[k] + 1] = -1;
					}
				}
				regex_push(&stack, alternate);
			}
			if (body && (!exit || i->greedy)) {
				state.pc++;
				state.repeats[i->x].start = state.position;
				for (size_t k = 0; k < i->id_count; k++) {
					state.captures[2 * i->ids[k]] = -1;
					state.captures[2 * i->ids[k] + 1] = -1;
				}
			} else
				state.pc = (size_t)i->y;
			continue;
		}
		case 10: {
			regex_repeat *reg = &state.repeats[i->x];
			if (reg->count >= i->minimum && reg->start == state.position)
				failed = true;
			else {
				if (reg->count == UINT64_MAX) {
					static const char message[] = "regexp counter overflow";
					adamic_panic(message, sizeof message - 1);
				}
				reg->count++;
				state.pc = (size_t)i->y;
				continue;
			}
			break;
		}
		default:
			adamic_unreachable();
		}
		if (failed) {
			if (stack == NULL)
				break;
			regex_free(state);
			regex_state *frame = stack;
			state = *frame;
			stack = frame->previous;
		} else
			state.pc++;
	}
finished:
	if (matched) {
		*position = state.position;
		memcpy(captures, state.captures, 2 * (p->captures + 1) * sizeof *captures);
	}
	regex_free(state);
	while (stack != NULL) {
		regex_state *frame = stack;
		stack = frame->previous;
		regex_free(*frame);
	}
	return matched;
}
static const char *const regex_names[] = {"__program", "lastIndex",	 "source",		"flags",
										  "global",	   "ignoreCase", "multiline",	"unicode",
										  "sticky",	   "hasIndices", "unicodeSets", "dotAll"};
static const bool regex_references[] = {false, false, true,	 true,	false, false,
										false, false, false, false, false, false};
static const adamic_shape regex_shape = {12, regex_names, regex_references, NULL};
adamic_object *adamic_regex_new(const adamic_regex_program *program, adamic_string *source,
								adamic_string *flags) {
	adamic_object *result = adamic_object_new(&regex_shape);
	result->slots[0].reference = (void *)program;
	result->slots[2].reference = adamic_retain(source);
	result->slots[3].reference = adamic_retain(flags);
	const unsigned bits[] = {8, 1, 2, 4, 16, 32, 64, 128};
	for (size_t k = 0; k < 8; k++)
		result->slots[4 + k].boolean = (program->flags & bits[k]) != 0;
	// unicode means u specifically, whereas the VM's bit 4 covers u and v.
	result->slots[7].boolean = (program->flags & 4) != 0 && !(program->flags & 64);
	return result;
}
static const adamic_regex_program *regex_program(adamic_object *regex) {
	return regex->slots[0].reference;
}
static size_t regex_to_length(double value) {
	if (!(value > 0))
		return 0;
	return (size_t)fmin(trunc(value), 9007199254740991.0);
}
static const uint16_t *regex_input(adamic_string *input, size_t *length) {
	const uint16_t *cached = adamic_string_utf16_view(input);
	if (cached != NULL) {
		*length = adamic_string_units(input);
		return cached;
	}
	// Decode immutable WTF-8 once, preserving lone surrogates and pair halves.
	// Repeated charCodeAt calls otherwise redo UTF-16-to-byte index lookups.
	uint16_t *units = regex_memory(input->length * sizeof *units);
	// Cached ASCII strings need only widening, not a branch at every byte. This
	// loop can be vectorized; the general WTF-8 decoder remains below.
	if (input->units == input->length + 1) {
		for (size_t at = 0; at < input->length; at++)
			units[at] = (unsigned char)input->bytes[at];
		*length = input->length;
		return units;
	}
	*length = 0;
	for (size_t at = 0; at < input->length;) {
		uint32_t point = (unsigned char)input->bytes[at++];
		if (point >= 128) {
			size_t extra = point < 0xe0 ? 1 : point < 0xf0 ? 2 : 3;
			point &= extra == 1 ? 0x1f : extra == 2 ? 0x0f : 0x07;
			while (extra-- != 0)
				point = (point << 6) | ((unsigned char)input->bytes[at++] & 0x3f);
		}
		if (point > 0xffff) {
			units[(*length)++] = (uint16_t)(0xd800 + ((point - 0x10000) >> 10));
			units[(*length)++] = (uint16_t)(0xdc00 + ((point - 0x10000) & 0x3ff));
		} else
			units[(*length)++] = (uint16_t)point;
	}
	return units;
}
static void regex_input_free(adamic_string *input, const uint16_t *units) {
	if (units != adamic_string_utf16_view(input))
		free((void *)units);
}
static size_t regex_advance(const uint16_t *units, size_t length, size_t at, bool unicode);
static inline __attribute__((always_inline)) ptrdiff_t *
regex_execute_kernel(adamic_object *regex, const uint16_t *input, size_t length, bool force_sticky,
					 uint64_t *steps, bool enhanced) {
	const adamic_regex_program *p = regex_program(regex);
	regex_workspace workspace;
	workspace.free_frames = NULL;
	workspace.used = 0;
	ADAMIC_CHECK_STACK();
	bool stateful = (p->flags & 24) != 0 || force_sticky;
	size_t start = stateful ? regex_to_length(regex->slots[1].number) : 0;
	size_t requested = start;
	// V8 rewinds an initial pair interior, then permits assertion-only paths
	// at that UTF-16 position if the rewound attempt failed. ECMA-262
	// 22.2.2.2/22.2.7.2 instead describe a code-point Input with no interior.
	if ((p->flags & 4) && start > 0 && start < length && high(input[start - 1]) && low(input[start]))
		start--;
	ptrdiff_t *captures = regex_memory(2 * (p->captures + 1) * sizeof *captures);
	while (start <= length) {
		ptrdiff_t at = (ptrdiff_t)start;
		if (p->anchored && at != 0)
			break;
		bool regular = enhanced && p->regular != NULL && regex_step_limit == 0 &&
					   regex_regular_mode != 0 &&
					   (regex_regular_mode == 2 || (length >= 2048 && p->fast_ascii == NULL));
		if (regular) {
			at = adamic_regex_regular_find(p, input, length, (size_t)at,
										   (p->flags & 16) || force_sticky);
			if (at < 0)
				break;
		}
		if (enhanced && p->before_count != 0 && !((p->flags & 16) || force_sticky)) {
			size_t scan = (size_t)at > p->before_count ? (size_t)at - p->before_count : 0;
			bool found_context = false;
			while (scan + p->before_count <= length) {
				const unsigned char *base = (const unsigned char *)(input + scan);
				const unsigned char *hit =
					memchr(base, (unsigned char)p->before[0], (length - scan) * sizeof *input);
				if (hit == NULL)
					break;
				scan += (size_t)(hit - base) / sizeof *input;
				if (scan + p->before_count <= length &&
					memcmp(input + scan, p->before, p->before_count * sizeof *input) == 0 &&
					scan + p->before_count >= (size_t)at) {
					at = (ptrdiff_t)(scan + p->before_count);
					found_context = true;
					break;
				}
				scan++;
			}
			if (!found_context)
				break;
		}
		for (size_t k = 0; k < 2 * (p->captures + 1); k++)
			captures[k] = -1;
		captures[0] = at;
		if (!regular && (!enhanced || p->before_count == 0) && !p->anchored &&
			!((p->flags & 16) || force_sticky)) {
			while ((size_t)at < length) {
				uint16_t c = input[at];
				bool possible = !p->filter_first || c >= 128 ||
								(p->first_ascii[c / 64] & (UINT64_C(1) << (c % 64)));
				if (possible &&
					(p->prefix_count == 0 ||
					 ((size_t)at + p->prefix_count <= length &&
					  memcmp(input + at, p->prefix, p->prefix_count * sizeof *input) == 0)))
					break;
				// memchr skips units lacking the first literal byte. Verify complete
				// units afterwards, so either byte order and false byte hits are safe.
				if (p->prefix_count && at + 1 < (ptrdiff_t)length) {
					const unsigned char *base = (const unsigned char *)(input + at + 1);
					const unsigned char *found = memchr(base, (unsigned char)p->prefix[0],
														(length - (size_t)at - 1) * sizeof *input);
					if (found == NULL) {
						at = (ptrdiff_t)length;
						break;
					}
					at += 1 + (ptrdiff_t)((found - base) / sizeof *input);
				} else {
					at = (ptrdiff_t)regex_advance(input, length, (size_t)at, (p->flags & 4) != 0);
				}
			}
			captures[0] = at;
		}
		if (p->fast != NULL && regex_step_limit == 0
				? p->fast(input, length, &at, captures, steps)
				: regex_run(p, input, length, &at, captures, steps, &workspace)) {
			captures[1] = at;
			if (stateful)
				regex->slots[1].number = (double)at;
			regex_workspace_destroy(&workspace);
			return captures;
		}
		if ((size_t)at >= length)
			break;
		if (p->anchored || (((p->flags & 16) || force_sticky) && start >= requested))
			break;
		// Irregexp can accept \B or (?!\W) inside a pair. Unlike the
		// code-point walk in 22.2.7.3 AdvanceStringIndex, try each code unit;
		// regex_read still rejects consuming a half of a Unicode pair.
		start = (size_t)at + 1;
	}
	if (stateful)
		regex->slots[1].number = 0;
	free(captures);
	regex_workspace_destroy(&workspace);
	return NULL;
}
// Compile separate kernels so short VM searches have no regular-engine or
// lookbehind-filter checks inside their per-candidate loop.
static __attribute__((noinline)) ptrdiff_t *regex_execute_vm(adamic_object *regex,
															 const uint16_t *input, size_t length,
															 bool force_sticky, uint64_t *steps) {
	return regex_execute_kernel(regex, input, length, force_sticky, steps, false);
}
static __attribute__((noinline)) ptrdiff_t *regex_execute_enhanced(adamic_object *regex,
																   const uint16_t *input,
																   size_t length, bool force_sticky,
																   uint64_t *steps) {
	return regex_execute_kernel(regex, input, length, force_sticky, steps, true);
}
static ptrdiff_t *regex_execute(adamic_object *regex, const uint16_t *input, size_t length,
								bool force_sticky, uint64_t *steps) {
	if (regex_step_limit == 0 && (length >= 2048 || regex_regular_mode == 2))
		return regex_execute_enhanced(regex, input, length, force_sticky, steps);
	return regex_execute_vm(regex, input, length, force_sticky, steps);
}
static const char *const match_names[] = {"index", "input", "groups", "indices"};
static const bool match_references[] = {false, true, true, true};
static const adamic_shape match_shape = {4, match_names, match_references, NULL};
static adamic_object *regex_groups(const adamic_regex_program *p, adamic_array *captures) {
	if (p->group_count == 0)
		return NULL;
	adamic_object *groups = adamic_object_new(p->group_shape);
	for (size_t k = 0; k < p->group_count; k++) {
		for (size_t j = 0; j < p->groups[k].count; j++) {
			void *value = captures->elements[p->groups[k].captures[j]].reference;
			if (value != NULL) {
				adamic_release(groups->slots[k].reference);
				groups->slots[k].reference = adamic_retain(value);
			}
		}
	}
	return groups;
}
static const char *const pair_names[] = {"0", "1"};
static const bool pair_references[] = {false, false};
static const adamic_shape pair_shape = {2, pair_names, pair_references, NULL};
static adamic_array *regex_result(const adamic_regex_program *p, adamic_string *input,
								  const ptrdiff_t *spans) {
	adamic_array *result = adamic_array_new(p->captures + 1, true);
	for (size_t k = 0; k <= p->captures; k++) {
		adamic_string *value =
			spans[2 * k] < 0
				? NULL
				: adamic_string_slice(input, (double)spans[2 * k], (double)spans[2 * k + 1], true);
		adamic_array_push(result, (adamic_value){.reference = value});
	}
	result->properties = adamic_object_new(&match_shape);
	result->properties->slots[0].number = (double)spans[0];
	result->properties->slots[1].reference = adamic_retain(input);
	result->properties->slots[2].reference = regex_groups(p, result);
	if (p->flags & 32) {
		adamic_array *indices = adamic_array_new(p->captures + 1, true);
		for (size_t k = 0; k <= p->captures; k++) {
			adamic_object *pair = NULL;
			if (spans[2 * k] >= 0) {
				pair = adamic_object_new(&pair_shape);
				pair->slots[0].number = (double)spans[2 * k];
				pair->slots[1].number = (double)spans[2 * k + 1];
			}
			adamic_array_push(indices, (adamic_value){.reference = pair});
		}
		indices->properties = adamic_object_new(&match_shape);
		indices->properties->slots[2].reference = regex_groups(p, indices);
		result->properties->slots[3].reference = indices;
	}
	return result;
}
adamic_maybe_boolean adamic_regex_done(adamic_object *object) {
	if (object == NULL) {
		static const char message[] = "undefined iterator result";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t k = 0; k < object->shape->count; k++)
		if (strcmp(object->shape->names[k], "done") == 0)
			return (adamic_maybe_boolean){true, object->slots[k].boolean};
	return (adamic_maybe_boolean){false, false};
}
void *adamic_regex_group_lookup(adamic_object *object, const char *name, bool optional) {
	if (object == NULL) {
		if (optional)
			return NULL;
		static const char message[] = "undefined named-group dictionary";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t k = 0; k < object->shape->count; k++)
		if (strcmp(name, object->shape->names[k]) == 0)
			return object->slots[k].reference;
	return NULL;
}
adamic_value adamic_regex_property(adamic_array *array, const char *name) {
	if (array == NULL) {
		static const char message[] = "null regular expression match result";
		adamic_panic(message, sizeof message - 1);
	}
	if (array->properties != NULL) {
		for (size_t k = 0; k < array->properties->shape->count; k++)
			if (strcmp(name, array->properties->shape->names[k]) == 0)
				return array->properties->slots[k];
	}
	return (adamic_value){.reference = NULL};
}
adamic_array *adamic_regex_exec(adamic_object *regex, adamic_string *input) {
	size_t length;
	const uint16_t *units = regex_input(input, &length);
	uint64_t steps = 0;
	ptrdiff_t *spans = regex_execute(regex, units, length, false, &steps);
	regex_input_free(input, units);
	if (spans == NULL)
		return NULL;
	adamic_array *result = regex_result(regex_program(regex), input, spans);
	free(spans);
	return result;
}
static __attribute__((noinline)) bool regex_test_general(adamic_object *regex, adamic_string *input,
														 size_t ascii_length) {
	const adamic_regex_program *p = regex_program(regex);
	ptrdiff_t first = -2;
	if (p->first_ascii_position != NULL && !(p->flags & 16) && regex_step_limit == 0 &&
		input->units == input->length + 1) {
		size_t start = (p->flags & 8) ? regex_to_length(regex->slots[1].number) : 0;
		first = p->first_ascii_position((const unsigned char *)input->bytes, ascii_length, start);
		if (first < 0) {
			if (p->flags & 8)
				regex->slots[1].number = 0;
			return false;
		}
	}

	if (p->regular != NULL && (p->fast_ascii == NULL || regex_regular_mode == 2) &&
		!(p->flags & 24) && regex_step_limit == 0 && regex_regular_mode != 0 &&
		(regex_regular_mode == 2 || ascii_length >= 2048) && input->units == input->length + 1) {
		int result =
			adamic_regex_regular_ascii(p, (const unsigned char *)input->bytes, ascii_length);
		if (result >= 0)
			return result != 0;
	}
	size_t length;
	const uint16_t *units = regex_input(input, &length);
	uint64_t steps = 0;
	ptrdiff_t *spans = regex_execute(regex, units, length, false, &steps);
	bool result = spans != NULL;
	free(spans);
	regex_input_free(input, units);
	return result;
}
bool adamic_regex_test(adamic_object *regex, adamic_string *input) {
	const adamic_regex_program *p = regex_program(regex);
	size_t ascii_length = input->units != 0 ? input->units - 1 : adamic_string_units(input);
	if (p->test_ascii != NULL && regex_step_limit == 0 && input->units == input->length + 1) {
		bool stateful = (p->flags & 24) != 0;
		ptrdiff_t at = stateful ? (ptrdiff_t)regex_to_length(regex->slots[1].number) : 0;
		// Reject absent prefixes before the indirect callback. Most literal
		// searches fail here, so they need no per-pattern call at all.
		if (p->test_prefix) {
			const unsigned char *bytes = (const unsigned char *)input->bytes;
			const unsigned char *found =
				(size_t)at > ascii_length
					? NULL
					: memchr(bytes + at, (unsigned char)p->prefix[0], ascii_length - (size_t)at);
			if (found == NULL) {
				if (stateful)
					regex->slots[1].number = 0;
				return false;
			}
			at = (ptrdiff_t)(found - bytes);
		}
		bool matched = p->test_ascii((const unsigned char *)input->bytes, ascii_length, &at);
		if (stateful)
			regex->slots[1].number = matched ? (double)at : 0;
		return matched;
	}
	return regex_test_general(regex, input, ascii_length);
}
static size_t regex_advance(const uint16_t *units, size_t length, size_t at, bool unicode) {
	return unicode && at + 1 < length && high(units[at]) && low(units[at + 1]) ? at + 2 : at + 1;
}
adamic_array *adamic_regex_match(adamic_string *input, adamic_object *regex) {
	const adamic_regex_program *p = regex_program(regex);
	if (!(p->flags & 8))
		return adamic_regex_exec(regex, input);
	regex->slots[1].number = 0;
	adamic_array *result = adamic_array_new(0, true);
	size_t length;
	const uint16_t *units = regex_input(input, &length);
	uint64_t steps = 0;
	for (;;) {
		ptrdiff_t *spans = regex_execute(regex, units, length, false, &steps);
		if (spans == NULL)
			break;
		adamic_array_push(result,
						  (adamic_value){.reference = adamic_string_slice(input, (double)spans[0],
																		  (double)spans[1], true)});
		if (spans[0] == spans[1])
			regex->slots[1].number = (double)regex_advance(
				units, length, regex_to_length(regex->slots[1].number), p->flags & 4);
		free(spans);
	}
	regex_input_free(input, units);
	if (result->length == 0) {
		adamic_release(result);
		return NULL;
	}
	return result;
}
static const char *const iterator_names[] = {"regex", "input", "done"};
static const bool iterator_references[] = {true, true, false};
static const adamic_shape iterator_shape = {3, iterator_names, iterator_references, NULL};
static void regex_require_global(const adamic_regex_program *p, const char *message) {
	if (!(p->flags & 8))
		adamic_panic(message, strlen(message));
}
adamic_object *adamic_regex_match_all(adamic_string *input, adamic_object *regex) {
	const adamic_regex_program *p = regex_program(regex);
	regex_require_global(
		p, "TypeError: String.prototype.matchAll called with a non-global RegExp argument");
	adamic_object *copy = adamic_regex_new(p, regex->slots[2].reference, regex->slots[3].reference);
	copy->slots[1].number = (double)regex_to_length(regex->slots[1].number);
	adamic_object *iterator = adamic_object_new(&iterator_shape);
	iterator->slots[0].reference = copy;
	iterator->slots[1].reference = adamic_retain(input);
	return iterator;
}
adamic_array *adamic_regex_iterator_step(adamic_object *iterator) {
	if (iterator->slots[2].boolean)
		return NULL;
	adamic_object *regex = iterator->slots[0].reference;
	adamic_string *input = iterator->slots[1].reference;
	adamic_array *result = adamic_regex_exec(regex, input);
	if (result == NULL)
		iterator->slots[2].boolean = true;
	else {
		adamic_string *whole = result->elements[0].reference;
		if (adamic_string_length(whole) == 0) {
			size_t length;
			const uint16_t *units = regex_input(input, &length);
			regex->slots[1].number =
				(double)regex_advance(units, length, regex_to_length(regex->slots[1].number),
									  regex_program(regex)->flags & 4);
			regex_input_free(input, units);
		}
	}
	return result;
}
static const char *const next_names[] = {"done", "value"};
static const bool next_references[] = {false, true};
static const adamic_shape next_shape = {2, next_names, next_references, NULL};
adamic_object *adamic_regex_next(adamic_object *iterator) {
	adamic_array *value = adamic_regex_iterator_step(iterator);
	adamic_object *result = adamic_object_new(&next_shape);
	result->slots[0].boolean = value == NULL;
	result->slots[1].reference = value;
	return result;
}
double adamic_regex_search(adamic_string *input, adamic_object *regex) {
	double previous = regex->slots[1].number;
	regex->slots[1].number = 0;
	size_t length;
	const uint16_t *units = regex_input(input, &length);
	uint64_t steps = 0;
	ptrdiff_t *spans = regex_execute(regex, units, length, false, &steps);
	double result = spans == NULL ? -1 : (double)spans[0];
	free(spans);
	regex_input_free(input, units);
	regex->slots[1].number = previous;
	return result;
}
adamic_array *adamic_regex_split(adamic_string *input, adamic_object *regex, double limit_value, bool default_limit) {
	adamic_array *result = adamic_array_new(0, true);
	uint32_t limit = (uint32_t)adamic_shift_right_unsigned(limit_value, 0);
	if (limit == 0)
		return result;
	const adamic_regex_program *p = regex_program(regex);
	adamic_object *copy = adamic_regex_new(p, regex->slots[2].reference, regex->slots[3].reference);
	size_t length;
	const uint16_t *units = regex_input(input, &length);
	uint64_t steps = 0;
	size_t previous = 0, at = 0;
	if (length == 0) {
		ptrdiff_t *spans = regex_execute(copy, units, length, true, &steps);
		if (spans == NULL)
			adamic_array_push(result, (adamic_value){.reference = adamic_retain(input)});
		free(spans);
		goto finished_split;
	}
	while (at < length) {
		copy->slots[1].number = (double)at;
		ptrdiff_t *spans = regex_execute(copy, units, length, true, &steps);
		if (spans == NULL) {
			// V8's split fast path scans UTF-16 candidates, unlike ECMA-262
			// 22.2.6.14's AdvanceStringIndex. A HeapNumber limit uses its generic
			// path under u; v retains code-unit candidate scanning there too.
			at = default_limit || (p->flags & 64) ? at + 1
				: regex_advance(units, length, at, p->flags & 4);
			continue;
		}
		size_t end = (size_t)spans[1];
		if (end == previous) {
			free(spans);
			at = regex_advance(units, length, at, p->flags & 4);
			continue;
		}
		adamic_array_push(result, (adamic_value){.reference = adamic_string_slice(
													 input, (double)previous, (double)at, true)});
		previous = end;
		if (result->length == limit) {
			free(spans);
			goto finished_split;
		}
		for (size_t k = 1; k <= p->captures; k++) {
			adamic_array_push(
				result, (adamic_value){.reference = spans[2 * k] < 0
														? NULL
														: adamic_string_slice(
															  input, (double)spans[2 * k],
															  (double)spans[2 * k + 1], true)});
			if (result->length == limit) {
				free(spans);
				goto finished_split;
			}
		}
		free(spans);
		at = previous;
	}
	adamic_array_push(result, (adamic_value){.reference = adamic_string_slice(
												 input, (double)previous, (double)length, true)});
finished_split:
	regex_input_free(input, units);
	adamic_release(copy);
	return result;
}
static void regex_piece(adamic_array *pieces, adamic_string *input, size_t first, size_t last) {
	adamic_array_push(pieces, (adamic_value){.reference = adamic_string_slice(input, (double)first,
																			  (double)last, true)});
}
static void regex_substitution(adamic_array *pieces, adamic_string *input, size_t input_length,
							   adamic_string *replacement, const uint16_t *text, size_t length,
							   const ptrdiff_t *spans, const adamic_regex_program *p) {
	size_t plain = 0;
	for (size_t at = 0; at < length; at++) {
		if (text[at] != '$' || at + 1 == length)
			continue;
		size_t consumed = 2;
		ptrdiff_t first = -1, last = -1;
		bool special = true;
		uint16_t next = text[at + 1];
		switch (next) {
		case '$':
			first = (ptrdiff_t)at;
			last = (ptrdiff_t)at + 1;
			break;
		case '&':
			first = spans[0];
			last = spans[1];
			break;
		case '`':
			first = 0;
			last = spans[0];
			break;
		case '\'':
			first = spans[1];
			last = (ptrdiff_t)input_length;
			break;
		case '<': {
			if (p->group_count == 0) {
				special = false;
				break;
			}
			size_t end = at + 2;
			while (end < length && text[end] != '>')
				end++;
			if (end == length) {
				special = false;
				break;
			}
			consumed = end - at + 1;
			// Names are compared as strings, retaining UTF-16 spelling for non-ASCII.
			adamic_string *name =
				adamic_string_slice(replacement, (double)(at + 2), (double)end, true);
			for (size_t k = 0; k < p->group_count; k++) {
				size_t bytes = strlen(p->groups[k].name);
				if (name->length == bytes && memcmp(name->bytes, p->groups[k].name, bytes) == 0) {
					for (size_t j = 0; j < p->groups[k].count; j++) {
						size_t id = p->groups[k].captures[j];
						if (spans[2 * id] >= 0) {
							first = spans[2 * id];
							last = spans[2 * id + 1];
						}
					}
				}
			}
			adamic_release(name);
			break;
		}
		default: {
			if (next < '0' || next > '9') {
				special = false;
				break;
			}
			size_t id = next - '0';
			if (at + 2 < length && text[at + 2] >= '0' && text[at + 2] <= '9') {
				size_t pair = id * 10 + text[at + 2] - '0';
				if (pair > 0 && pair <= p->captures) {
					id = pair;
					consumed = 3;
				}
			}
			if (id == 0 || id > p->captures) {
				special = false;
				break;
			}
			first = spans[2 * id];
			last = spans[2 * id + 1];
			break;
		}
		}
		if (!special)
			continue;
		regex_piece(pieces, replacement, plain, at);
		if (first >= 0)
			regex_piece(pieces, next == '$' ? replacement : input, (size_t)first, (size_t)last);
		at += consumed - 1;
		plain = at + 1;
	}
	regex_piece(pieces, replacement, plain, length);
}
adamic_string *adamic_regex_replace(adamic_string *input, adamic_object *regex,
									adamic_string *replacement, bool require_global) {
	const adamic_regex_program *p = regex_program(regex);
	if (require_global)
		regex_require_global(
			p, "TypeError: String.prototype.replaceAll called with a non-global RegExp argument");
	bool global = (p->flags & 8) != 0;
	if (global)
		regex->slots[1].number = 0;
	size_t length, replacement_length;
	const uint16_t *units = regex_input(input, &length),
			 *text = regex_input(replacement, &replacement_length);
	uint64_t steps = 0;
	adamic_array *pieces = adamic_array_new(0, true);
	size_t previous = 0;
	for (;;) {
		ptrdiff_t *spans = regex_execute(regex, units, length, false, &steps);
		if (spans == NULL)
			break;
		regex_piece(pieces, input, previous, (size_t)spans[0]);
		regex_substitution(pieces, input, length, replacement, text, replacement_length, spans, p);
		previous = (size_t)spans[1];
		bool empty = spans[0] == spans[1];
		free(spans);
		if (!global)
			break;
		if (empty)
			regex->slots[1].number = (double)regex_advance(
				units, length, regex_to_length(regex->slots[1].number), p->flags & 4);
	}
	regex_piece(pieces, input, previous, length);
	adamic_string *result = adamic_array_join(pieces, &adamic_string_empty, adamic_join_strings);
	adamic_release(pieces);
	regex_input_free(replacement, text);
	regex_input_free(input, units);
	return result;
}
