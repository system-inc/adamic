#ifndef ADAMIC_REGEXP_COMPILE_SETS_H
#define ADAMIC_REGEXP_COMPILE_SETS_H
#include "adamic.h"
#include "regexp_compile_parser.h"
/* Temporary sets belong to the parse arena. Emitted instruction arrays borrow
 * their normalized ranges and ordered strings for the arena's lifetime. */
typedef struct {
    adamic_regex_range *ranges;
    size_t range_count, range_capacity;
    adamic_regex_text *strings;
    size_t string_count, string_capacity;
} adamic_regex_compile_set;
adamic_regex_compile_set adamic_regex_compile_character(adamic_regex_parse_result *, const adamic_regex_parse_node *, unsigned);
adamic_regex_compile_set adamic_regex_compile_class(adamic_regex_parse_result *, const adamic_regex_parse_node *, unsigned);
adamic_regex_compile_set adamic_regex_compile_fold(adamic_regex_parse_result *, adamic_regex_compile_set, unsigned);
adamic_regex_compile_set adamic_regex_compile_negate(adamic_regex_parse_result *, adamic_regex_compile_set, unsigned);
void adamic_regex_compile_add_range(adamic_regex_parse_result *, adamic_regex_compile_set *, uint32_t, uint32_t);
adamic_regex_compile_set adamic_regex_compile_matched(adamic_regex_parse_result *, adamic_regex_compile_set, unsigned);
adamic_regex_compile_set adamic_regex_compile_union(adamic_regex_parse_result *, adamic_regex_compile_set, adamic_regex_compile_set);
adamic_regex_compile_set adamic_regex_compile_combine(adamic_regex_parse_result *, adamic_regex_compile_set, adamic_regex_compile_set, bool);
bool adamic_regex_compile_same(adamic_regex_compile_set, adamic_regex_compile_set);
#endif
