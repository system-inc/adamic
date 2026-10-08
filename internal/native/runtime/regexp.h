#ifndef ADAMIC_REGEXP_H
#define ADAMIC_REGEXP_H

// Kept in the same opcode order as internal/regexp/matcher.go.
typedef struct {
	uint32_t first, last;
} adamic_regex_range;
typedef struct {
	const uint32_t *points;
	size_t count;
} adamic_regex_text;
typedef struct {
	const char *name;
	const size_t *captures;
	size_t count;
} adamic_regex_group;
typedef struct adamic_regex_program adamic_regex_program;
typedef struct {
	int op, x, y, direction;
	unsigned flags;
	int assertion;
	bool negative, greedy, unbounded;
	uint64_t minimum, maximum;
	const adamic_regex_range *ranges;
	size_t range_count;
	const adamic_regex_text *strings;
	size_t string_count;
	const size_t *ids;
	size_t id_count;
	const adamic_regex_program *look;
} adamic_regex_instruction;
typedef struct {
	int op, x, y;
	size_t source;
} adamic_regex_regular_instruction;
typedef struct {
	const adamic_regex_regular_instruction *code;
	size_t count, start;
	const unsigned char *classes, *characters;
	size_t class_count;
	const uint64_t *masks;
	size_t words;
} adamic_regex_regular;
struct adamic_regex_program {
	const adamic_regex_instruction *code;
	bool (*test_ascii)(const unsigned char *, size_t, ptrdiff_t *);
	bool test_prefix;
	size_t captures, repeats;
	unsigned flags;
	const adamic_regex_group *groups;
	size_t group_count;
	const adamic_shape *group_shape;
	uint64_t first_ascii[2];
	bool filter_first;
	const uint16_t *prefix;
	size_t prefix_count;
	bool (*fast)(const uint16_t *, size_t, ptrdiff_t *, ptrdiff_t *, uint64_t *);
	bool anchored;
	bool (*fast_ascii)(const unsigned char *, size_t, ptrdiff_t *, ptrdiff_t *, uint64_t *);
	const adamic_regex_regular *regular;
	const uint16_t *before;
	size_t before_count;
	ptrdiff_t (*first_ascii_position)(const unsigned char *, size_t, size_t);
};

adamic_object *adamic_regex_new(const adamic_regex_program *program, adamic_string *source,
								adamic_string *flags);
adamic_object *adamic_regex_new_owned(const adamic_regex_program *program, adamic_string *source,
                                    adamic_string *flags, adamic_array *storage);
bool adamic_regex_test(adamic_object *regex, adamic_string *input);
adamic_array *adamic_regex_exec(adamic_object *regex, adamic_string *input);
adamic_array *adamic_regex_match(adamic_string *input, adamic_object *regex);
adamic_object *adamic_regex_match_all(adamic_string *input, adamic_object *regex);
adamic_array *adamic_regex_iterator_step(adamic_object *iterator);
adamic_object *adamic_regex_next(adamic_object *iterator);
adamic_string *adamic_regex_replace(adamic_string *input, adamic_object *regex,
									adamic_string *replacement, bool require_global);
adamic_array *adamic_regex_split(adamic_string *input, adamic_object *regex, double limit, bool default_limit);
double adamic_regex_search(adamic_string *input, adamic_object *regex);
adamic_value adamic_regex_property(adamic_array *array, const char *name);
adamic_maybe_boolean adamic_regex_done(adamic_object *object);
void *adamic_regex_group_lookup(adamic_object *object, const char *name, bool optional);
void adamic_regex_set_step_limit(uint64_t limit);
void adamic_regex_set_regular_enabled(bool enabled);
void adamic_regex_set_regular_mode(int mode);
ptrdiff_t adamic_regex_regular_find(const adamic_regex_program *, const uint16_t *, size_t, size_t,
									bool);
int adamic_regex_regular_ascii(const adamic_regex_program *, const unsigned char *, size_t);
bool adamic_regex_read(const uint16_t *, size_t, ptrdiff_t, int, bool, uint32_t *, ptrdiff_t *);
uint32_t adamic_regex_canonical(uint32_t, unsigned);
bool adamic_regex_word(uint32_t, unsigned);
bool adamic_regex_contains(const adamic_regex_instruction *, uint32_t);
typedef struct { size_t argument; bool rest; const char *name; bool optional; } adamic_regex_replacement_group;
adamic_string *adamic_regex_replace_callback(adamic_string *input, adamic_object *regex,
 adamic_closure *callback, bool require_global, const unsigned char *parameters,
 size_t parameter_count, unsigned rest_kind, unsigned return_kind,
 const adamic_regex_replacement_group *group_checks, size_t group_check_count);
#endif
