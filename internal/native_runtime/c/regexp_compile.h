#ifndef ADAMIC_REGEXP_COMPILE_H
#define ADAMIC_REGEXP_COMPILE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

/* Compiler tables are independent of the matcher and contain no pointers into Go. */
typedef struct { uint32_t first, last; } adamic_regex_compile_range;
typedef struct { const uint32_t *points; size_t count; } adamic_regex_compile_text;
typedef struct {
    const adamic_regex_compile_range *ranges;
    size_t range_count;
    const adamic_regex_compile_text *strings;
    size_t string_count;
    bool string_property;
} adamic_regex_compile_property;

/* Counted input is required: JavaScript strings can contain embedded NUL. */
const adamic_regex_compile_property *adamic_regex_compile_lookup_property(
    const unsigned char *expression, size_t length, bool unicode_sets);
#endif
