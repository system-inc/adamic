#ifndef ADAMIC_REGEXP_COMPILE_PARSER_H
#define ADAMIC_REGEXP_COMPILE_PARSER_H
#include "regexp_compile.h"

enum {
    REGEX_PARSE_DISJUNCTION, REGEX_PARSE_ALTERNATIVE, REGEX_PARSE_CHARACTER,
    REGEX_PARSE_ASSERTION, REGEX_PARSE_DOT, REGEX_PARSE_QUANTIFIER,
    REGEX_PARSE_GROUP, REGEX_PARSE_REFERENCE, REGEX_PARSE_CLASS,
    REGEX_PARSE_NEGATION, REGEX_PARSE_UNION, REGEX_PARSE_INTERSECTION,
    REGEX_PARSE_SUBTRACTION, REGEX_PARSE_RANGE, REGEX_PARSE_CLASS_CHARACTER,
    REGEX_PARSE_CLASS_STRING
};
enum { REGEX_FLAG_I=1, REGEX_FLAG_M=2, REGEX_FLAG_U=4, REGEX_FLAG_G=8,
       REGEX_FLAG_Y=16, REGEX_FLAG_D=32, REGEX_FLAG_V=64, REGEX_FLAG_S=128 };
/* Lists preserve the Go tree's source order. Decimal bounds remain arbitrary
 * precision until the bytecode emitter applies its uint64 refusal. */
typedef struct adamic_regex_parse_node adamic_regex_parse_node;
struct adamic_regex_parse_node {
    int type, kind;
    uint32_t value;
    unsigned enable, disable;
    bool negated, greedy;
    size_t start, end, index;
    const char *name, *minimum, *maximum;
    adamic_regex_parse_node *children, *next, *left, *right;
};
typedef struct adamic_regex_parse_memory adamic_regex_parse_memory;
typedef struct {
    unsigned flags;
    size_t captures;
    adamic_regex_parse_node *body;
    const unsigned char *source;
    size_t length;
    /* status: 0 accepted, 1 Node SyntaxError, 2 allocation failure,
     * 3 V8 divergence, 4 unsupported native feature. Only 1 is catchable. */
    int status;
    size_t error_offset;
    const char *reference_reason, *node_reason, *message;
    size_t message_length;
    adamic_regex_parse_memory *memory;
} adamic_regex_parse_result;
void adamic_regex_parse(const unsigned char *pattern, size_t length,
    const unsigned char *flags, size_t flag_length, adamic_regex_parse_result *result);
/* Node-compatible production parser; the reference entry retains exact MVs. */
void adamic_regex_parse_native(const unsigned char *pattern, size_t length,
    const unsigned char *flags, size_t flag_length, adamic_regex_parse_result *result);
/* Transfer malloc blocks to a counted owner; the callback owns each block. */
void adamic_regex_parse_take_memory(adamic_regex_parse_result *result, void (*take)(void *, void *), void *owner);
void adamic_regex_parse_free(adamic_regex_parse_result *result);
/* Compiler temporaries share this result's allocation ownership. */
void *adamic_regex_parse_allocate(adamic_regex_parse_result *result, size_t size);
#endif
