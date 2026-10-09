#include "regexp_compile.h"
// The ordinary runtime is linked with --whole-archive. Keep compiler data
// out until a dynamic RegExp build explicitly opts in.
#ifdef ADAMIC_REGEXP_RUNTIME_COMPILER
#include "regexp_compile_tables.h"

/* Compare counted WTF-8 input with the generated ASCII alias. */
static int regex_compile_compare(const unsigned char *text, size_t length, const char *alias) {
    size_t at = 0;
    while (at < length && alias[at] != '\0') {
        unsigned char byte = (unsigned char)alias[at];
        if (text[at] != byte) return text[at] < byte ? -1 : 1;
        at++;
    }
    if (at < length) return 1;
    return alias[at] == '\0' ? 0 : -1;
}

const adamic_regex_compile_property *adamic_regex_compile_lookup_property(
    const unsigned char *expression, size_t length, bool unicode_sets) {
    size_t low = 0, high = sizeof(regex_compile_aliases) / sizeof(regex_compile_aliases[0]);
    while (low < high) {
        size_t middle = low + (high - low) / 2;
        int order = regex_compile_compare(expression, length, regex_compile_aliases[middle].name);
        if (order > 0) low = middle + 1;
        else high = middle;
    }
    if (low == sizeof(regex_compile_aliases) / sizeof(regex_compile_aliases[0]) ||
        regex_compile_compare(expression, length, regex_compile_aliases[low].name) != 0) return NULL;
    const adamic_regex_compile_property *property = regex_compile_aliases[low].property;
    if (property->string_property && !unicode_sets) return NULL;
    return property;
}

#endif
