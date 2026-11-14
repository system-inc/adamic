// Capture-free regular search. The ordered VM recovers the selected match.
#include "adamic.h"
#include <string.h>

#define REGULAR_LIMIT 768
#define REGULAR_WORDS 12
#define DFA_LIMIT 128

typedef struct {
	uint64_t bits[REGULAR_WORDS];
} regular_bits;
static void regular_put(regular_bits *set, size_t pc) {
	set->bits[pc / 64] |= UINT64_C(1) << (pc % 64);
}
static int regular_take(regular_bits *set, size_t words) {
	for (size_t k = 0; k < words; k++)
		if (set->bits[k]) {
			int bit = __builtin_ctzll(set->bits[k]);
			set->bits[k] &= set->bits[k] - 1;
			return (int)(k * 64) + (int)bit;
		}
	return -1;
}
static bool regular_line(uint32_t c) { return c == 10 || c == 13 || c == 0x2028 || c == 0x2029; }
static bool regular_assert(const adamic_regex_instruction *i, const uint16_t *input, size_t length,
						   ptrdiff_t at) {
	if (i->assertion == 0)
		return at == 0 || ((i->flags & 2) && regular_line(input[at - 1]));
	if (i->assertion == 1)
		return (size_t)at == length || ((i->flags & 2) && regular_line(input[at]));
	uint32_t left = 0, right = 0;
	ptrdiff_t next;
	bool l = adamic_regex_read(input, length, at, -1, i->flags & 4, &left, &next),
		 r = adamic_regex_read(input, length, at, 1, i->flags & 4, &right, &next);
	bool boundary =
		(l && adamic_regex_word(left, i->flags)) != (r && adamic_regex_word(right, i->flags));
	return boundary != (i->assertion == 3);
}
static void regular_origin(regular_bits *pending, regular_bits *active, ptrdiff_t *origins,
						   size_t pc, ptrdiff_t origin) {
	if (origins[pc] < 0 || origin < origins[pc]) {
		origins[pc] = origin;
		regular_put(pending, pc);
		regular_put(active, pc);
	}
}
ptrdiff_t adamic_regex_regular_find(const adamic_regex_program *p, const uint16_t *input,
									size_t length, size_t start, bool sticky) {
	ADAMIC_CHECK_STACK();
	const adamic_regex_regular *r = p->regular;
	ptrdiff_t first[REGULAR_LIMIT], second[REGULAR_LIMIT], *current = first, *next = second;
	for (size_t k = 0; k < r->count; k++)
		current[k] = next[k] = -1;
	regular_bits active = {{0}}, following = {{0}};
	ptrdiff_t best = -1;
	for (ptrdiff_t at = (ptrdiff_t)start;;) {
		regular_bits pending = active;
		if (best < 0 && ((size_t)at == start || (!sticky && !p->anchored)))
			regular_origin(&pending, &active, current, r->start, at);
		int pc;
		while ((pc = regular_take(&pending, r->words)) >= 0) {
			ptrdiff_t origin = current[pc];
			if (best >= 0 && origin >= best)
				continue;
			const adamic_regex_regular_instruction *n = &r->code[pc];
			if (n->op == 0)
				best = origin;
			else if (n->op == 2) {
				regular_origin(&pending, &active, current, (size_t)n->x, origin);
				regular_origin(&pending, &active, current, (size_t)n->y, origin);
			} else if (n->op == 5 && regular_assert(&p->code[n->source], input, length, at))
				regular_origin(&pending, &active, current, (size_t)n->x, origin);
		}
		if ((size_t)at == length)
			return best;
		uint32_t c;
		ptrdiff_t to;
		bool readable = adamic_regex_read(input, length, at, 1, p->flags & 4, &c, &to);
		if (!readable)
			to = at + 1;
		while ((pc = regular_take(&active, r->words)) >= 0) {
			ptrdiff_t origin = current[pc];
			if (origin < 0 || (best >= 0 && origin >= best))
				continue;
			const adamic_regex_regular_instruction *n = &r->code[pc];
			if (n->op == 1 && readable) {
				const adamic_regex_instruction *i = &p->code[n->source];
				if (adamic_regex_contains(i, adamic_regex_canonical(c, i->flags)) &&
					(next[n->x] < 0 || origin < next[n->x])) {
					next[n->x] = origin;
					regular_put(&following, (size_t)n->x);
				}
			}
		}
		bool alive = false;
		for (size_t k = 0; k < r->words; k++)
			alive = alive || following.bits[k] != 0;
		if (!alive && (best >= 0 || sticky || p->anchored))
			return best;
		ptrdiff_t *swap = current;
		current = next;
		next = swap;
		for (size_t k = 0; k < r->count; k++)
			next[k] = -1;
		active = following;
		memset(&following, 0, sizeof following);
		at = to;
	}
}

// Each call owns a bounded lazy DFA cache. No shared mutation, heap allocation,
// lifetime hook or collector is needed. Cache exhaustion falls back to the VM.
// A state is an NFA subset plus previous-character assertion context. ASCII
// alphabet classes make a hot transition one byte lookup and one table lookup.
typedef struct {
	regular_bits raw;
	unsigned context;
	uint16_t transitions[128];
} regular_dfa_state;
static bool regular_ascii_assert(const adamic_regex_instruction *i, unsigned context, int c) {
	bool end = c < 0, word = !end && adamic_regex_word((uint32_t)c, 0),
		 line = !end && regular_line((uint32_t)c);
	if (i->assertion == 0)
		return (context & 4) || ((i->flags & 2) && (context & 2));
	if (i->assertion == 1)
		return end || ((i->flags & 2) && line);
	bool boundary = ((context & 1) != 0) != word;
	return boundary != (i->assertion == 3);
}
static bool regular_closure(const adamic_regex_program *p, regular_bits raw, unsigned context,
							int c, regular_bits *consuming) {
	const adamic_regex_regular *r = p->regular;
	regular_bits pending = raw, seen = {{0}};
	int pc;
	while ((pc = regular_take(&pending, r->words)) >= 0) {
		uint64_t bit = UINT64_C(1) << (pc % 64);
		if (seen.bits[pc / 64] & bit)
			continue;
		seen.bits[pc / 64] |= bit;
		const adamic_regex_regular_instruction *n = &r->code[pc];
		if (n->op == 0)
			return true;
		if (n->op == 1)
			regular_put(consuming, (size_t)pc);
		else if (n->op == 2) {
			regular_put(&pending, (size_t)n->x);
			regular_put(&pending, (size_t)n->y);
		} else if (n->op == 5 && regular_ascii_assert(&p->code[n->source], context, c))
			regular_put(&pending, (size_t)n->x);
	}
	return false;
}
int adamic_regex_regular_ascii(const adamic_regex_program *p, const unsigned char *input,
							   size_t length) {
	ADAMIC_CHECK_STACK();
	const adamic_regex_regular *r = p->regular;
	regular_dfa_state cache[DFA_LIMIT];
	size_t count = 1, state = 0;
	memset(&cache[0], 0, sizeof cache[0]);
	regular_put(&cache[0].raw, r->start);
	cache[0].context = 4;
	for (size_t at = 0; at < length; at++) {
		unsigned cls = r->classes[input[at]];
		uint16_t target = cache[state].transitions[cls];
		if (target == UINT16_MAX)
			return 1;
		if (target == 0) {
			regular_bits consuming = {{0}}, next = {{0}};
			int c = r->characters[cls];
			if (regular_closure(p, cache[state].raw, cache[state].context, c, &consuming)) {
				cache[state].transitions[cls] = UINT16_MAX;
				return 1;
			}
			const uint64_t *mask = r->masks + cls * r->words;
			for (size_t k = 0; k < r->words; k++)
				consuming.bits[k] &= mask[k];
			int pc;
			while ((pc = regular_take(&consuming, r->words)) >= 0)
				regular_put(&next, (size_t)r->code[pc].x);
			if (!p->anchored)
				regular_put(&next, r->start);
			unsigned context = (adamic_regex_word((uint32_t)c, 0) ? 1u : 0u) |
							   (regular_line((uint32_t)c) ? 2u : 0u);
			size_t found = 0;
			while (found < count &&
				   (cache[found].context != context ||
					memcmp(cache[found].raw.bits, next.bits, r->words * sizeof(uint64_t)) != 0))
				found++;
			if (found == count) {
				if (count == DFA_LIMIT)
					return -1;
				memset(&cache[count], 0, sizeof cache[count]);
				cache[count].raw = next;
				cache[count].context = context;
				count++;
			}
			target = (uint16_t)(found + 1);
			cache[state].transitions[cls] = target;
		}
		state = target - 1;
	}
	regular_bits consuming = {{0}};
	return regular_closure(p, cache[state].raw, cache[state].context, -1, &consuming) ? 1 : 0;
}
