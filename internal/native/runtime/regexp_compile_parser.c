#include "regexp_compile_parser.h"
#ifdef ADAMIC_REGEXP_RUNTIME_COMPILER
#include <stdlib.h>
#include <string.h>

/* A parse owns every allocation, on success and on every error path. No AST
 * node is recursively freed, so dropping a deep tree needs no C stack. */
struct adamic_regex_parse_memory {
    adamic_regex_parse_memory *next;
    max_align_t alignment;
};
typedef adamic_regex_parse_node regex_node;
typedef struct {
    adamic_regex_parse_result *result;
    size_t position, capture_count;
    bool named_capture;
    regex_node *pending;
} regex_parser;

static void *regex_parse_allocate(regex_parser *p, size_t size) {
    if (size > SIZE_MAX - sizeof(adamic_regex_parse_memory)) {
        p->result->status = 2;
        return NULL;
    }
    adamic_regex_parse_memory *memory = calloc(1, sizeof(*memory) + size);
    if (memory == NULL) { p->result->status = 2; return NULL; }
    memory->next = p->result->memory;
    p->result->memory = memory;
    return memory + 1;
}
void *adamic_regex_parse_allocate(adamic_regex_parse_result *result, size_t size) {
    regex_parser parser = {.result = result};
    return regex_parse_allocate(&parser, size);
}
void adamic_regex_parse_free(adamic_regex_parse_result *result) {
    adamic_regex_parse_memory *memory = result->memory;
    while (memory != NULL) {
        adamic_regex_parse_memory *next = memory->next;
        free(memory);
        memory = next;
    }
    memset(result, 0, sizeof(*result));
}
static regex_node *regex_parse_new(regex_parser *p, int type) {
    regex_node *node = regex_parse_allocate(p, sizeof(*node));
    if (node != NULL) node->type = type;
    return node;
}
static regex_node *regex_parse_fail(regex_parser *p, size_t at, const char *reference, const char *node) {
    if (p->result->status == 0) {
        p->result->status = 1;
        p->result->error_offset = at;
        p->result->reference_reason = reference;
        p->result->node_reason = node;
    }
    return NULL;
}
static bool regex_parse_unicode(regex_parser *p) { return (p->result->flags & REGEX_FLAG_U) != 0; }
static bool regex_parse_sets(regex_parser *p) { return (p->result->flags & REGEX_FLAG_V) != 0; }
static bool regex_parse_done(regex_parser *p) { return p->pending == NULL && p->position >= p->result->length; }
static unsigned char regex_parse_peek(regex_parser *p) {
    if (p->pending != NULL) return 0xff;
    return p->position >= p->result->length ? 0 : p->result->source[p->position];
}
static bool regex_parse_take(regex_parser *p, unsigned char byte) {
    if (regex_parse_peek(p) != byte) return false;
    p->position++;
    return true;
}
static bool regex_parse_match(regex_parser *p, const char *text) {
    size_t length = strlen(text);
    return p->position <= p->result->length && length <= p->result->length - p->position &&
        memcmp(p->result->source + p->position, text, length) == 0;
}
static bool regex_parse_in(unsigned char byte, const char *text) {
    return byte != 0 && strchr(text, byte) != NULL;
}
static bool regex_parse_digit(unsigned char byte) { return byte >= '0' && byte <= '9'; }
static int regex_parse_hex_digit(unsigned char byte) {
    if (byte >= '0' && byte <= '9') return byte - '0';
    if (byte >= 'a' && byte <= 'f') return byte - 'a' + 10;
    if (byte >= 'A' && byte <= 'F') return byte - 'A' + 10;
    return -1;
}
static uint32_t regex_parse_rune(regex_parser *p, size_t *width) {
    const unsigned char *s = p->result->source + p->position;
    size_t count = p->result->length - p->position;
    *width = 1;
    if (count == 0) return 0xfffd;
    if (s[0] < 0x80) return s[0];
    unsigned n = s[0] >= 0xf0 ? 4 : s[0] >= 0xe0 ? 3 : s[0] >= 0xc2 ? 2 : 0;
    if (n == 0 || n > count || s[0] > 0xf4) return 0xfffd;
    uint32_t point = s[0] & (0x7fU >> n);
    for (unsigned i = 1; i < n; i++) {
        if ((s[i] & 0xc0) != 0x80) return 0xfffd;
        point = (point << 6) | (s[i] & 63);
    }
    if ((n == 2 && point < 0x80) || (n == 3 && point < 0x800) ||
        (n == 4 && (point < 0x10000 || point > 0x10ffff))) return 0xfffd;
    /* WTF-8 surrogates are deliberately accepted. */
    *width = n;
    return point;
}
static regex_node *regex_parse_character(regex_parser *p, uint32_t value, int kind, size_t start) {
    regex_node *node = regex_parse_new(p, REGEX_PARSE_CHARACTER);
    if (node != NULL) { node->value = value; node->kind = kind; node->start = start; node->end = p->position; }
    return node;
}
static regex_node *regex_parse_literal(regex_parser *p) {
    if (p->pending != NULL) { regex_node *node = p->pending; p->pending = NULL; return node; }
    size_t start = p->position, width;
    uint32_t point = regex_parse_rune(p, &width);
    if (point == 0xfffd && width == 1) return regex_parse_fail(p, start, "invalid UTF-8", "Invalid escape");
    p->position += width;
    if (point > 0xffff && !regex_parse_unicode(p)) {
        uint32_t value = point - 0x10000;
        p->pending = regex_parse_character(p, 0xdc00 + value % 0x400, 0, start);
        point = 0xd800 + value / 0x400;
    }
    return regex_parse_character(p, point, 0, start);
}
static char *regex_parse_copy(regex_parser *p, size_t start, size_t end) {
    if (end < start || end - start == SIZE_MAX) { p->result->status = 2; return NULL; }
    char *text = regex_parse_allocate(p, end - start + 1);
    if (text != NULL) memcpy(text, p->result->source + start, end - start);
    return text;
}
static const char *regex_parse_decimal(regex_parser *p) {
    size_t start = p->position;
    while (!regex_parse_done(p) && regex_parse_digit(regex_parse_peek(p))) p->position++;
    if (start == p->position) return NULL;
    while (start + 1 < p->position && p->result->source[start] == '0') start++;
    return regex_parse_copy(p, start, p->position);
}
static int regex_parse_decimal_compare(const char *a, const char *b) {
    size_t al = strlen(a), bl = strlen(b);
    if (al != bl) return al < bl ? -1 : 1;
    return strcmp(a, b);
}
static bool regex_parse_quantifier_ahead(regex_parser *p) {
    unsigned char byte = regex_parse_peek(p);
    if (regex_parse_in(byte, "*+?")) return true;
    if (byte != '{') return false;
    size_t at = p->position + 1, length = p->result->length;
    const unsigned char *source = p->result->source;
    if (at >= length || !regex_parse_digit(source[at])) return false;
    while (at < length && regex_parse_digit(source[at])) at++;
    if (at < length && source[at] == ',') { at++; while (at < length && regex_parse_digit(source[at])) at++; }
    return at < length && source[at] == '}';
}
static bool regex_parse_hex_fixed(regex_parser *p, unsigned count, uint32_t *point) {
    if (p->position > p->result->length || count > p->result->length - p->position) return false;
    uint32_t value = 0;
    for (unsigned i = 0; i < count; i++) {
        int digit = regex_parse_hex_digit(p->result->source[p->position + i]);
        if (digit < 0) return false;
        value = (value << 4) | (unsigned)digit;
    }
    p->position += count;
    *point = value;
    return true;
}
static bool regex_parse_hex_braced(regex_parser *p, uint32_t *point) {
    size_t start = p->position;
    uint64_t value = 0;
    while (!regex_parse_done(p) && regex_parse_peek(p) != '}') {
        int digit = regex_parse_hex_digit(regex_parse_peek(p));
        if (digit < 0 || value > (0x10ffffU - (unsigned)digit) / 16) return false;
        value = value * 16 + (unsigned)digit;
        p->position++;
    }
    if (p->position == start || !regex_parse_take(p, '}')) return false;
    /* Accumulation above is bounded before multiplication on every target. */
    *point = (uint32_t)value;
    return true;
}
static regex_node *regex_parse_unicode_escape(regex_parser *p, size_t start) {
    uint32_t first;
    if (regex_parse_take(p, '{')) {
        if (!regex_parse_unicode(p)) { p->position = start + 2; return regex_parse_character(p, 'u', 1, start); }
        if (!regex_parse_hex_braced(p, &first)) return regex_parse_fail(p, start, "invalid Unicode escape", "Invalid Unicode escape");
    } else {
        if (!regex_parse_hex_fixed(p, 4, &first)) {
            if (!regex_parse_unicode(p)) { p->position = start + 2; return regex_parse_character(p, 'u', 1, start); }
            return regex_parse_fail(p, start, "invalid hex escape", "Invalid Unicode escape");
        }
        if (regex_parse_unicode(p) && first >= 0xd800 && first <= 0xdbff && regex_parse_match(p, "\\u")) {
            size_t second_start = p->position;
            uint32_t second;
            p->position += 2;
            if (regex_parse_hex_fixed(p, 4, &second) && second >= 0xdc00 && second <= 0xdfff)
                first = 0x10000 + (first - 0xd800) * 0x400 + second - 0xdc00;
            else p->position = second_start;
        }
    }
    return regex_parse_character(p, first, 1, start);
}
static bool regex_parse_property_contains(const char *name, uint32_t point) {
    const adamic_regex_compile_property *property = adamic_regex_compile_lookup_property((const unsigned char *)name, strlen(name), true);
    if (property == NULL) return false;
    size_t low = 0, high = property->range_count;
    while (low < high) {
        size_t middle = low + (high - low) / 2;
        if (property->ranges[middle].last < point) low = middle + 1;
        else high = middle;
    }
    return low < property->range_count && property->ranges[low].first <= point;
}
static bool regex_parse_identifier(uint32_t point, bool first) {
    if (point == '$' || point == '_') return true;
    if (!first && (point == 0x200c || point == 0x200d)) return true;
    return regex_parse_property_contains(first ? "ID_Start" : "ID_Continue", point);
}
static unsigned regex_parse_encode(uint32_t point, char *out) {
    if (point < 0x80) { out[0] = (char)point; return 1; }
    if (point < 0x800) { out[0] = (char)(0xc0 | point >> 6); out[1] = (char)(0x80 | (point & 63)); return 2; }
    if (point < 0x10000) { out[0] = (char)(0xe0 | point >> 12); out[1] = (char)(0x80 | ((point >> 6) & 63)); out[2] = (char)(0x80 | (point & 63)); return 3; }
    out[0] = (char)(0xf0 | point >> 18); out[1] = (char)(0x80 | ((point >> 12) & 63));
    out[2] = (char)(0x80 | ((point >> 6) & 63)); out[3] = (char)(0x80 | (point & 63)); return 4;
}
static const char *regex_parse_group_name(regex_parser *p) {
    size_t start = p->position, count = 0;
    if (p->result->length - start > (SIZE_MAX - 1) / 4) { p->result->status = 2; return NULL; }
    char *name = regex_parse_allocate(p, (p->result->length - start) * 4 + 1);
    if (name == NULL) return NULL;
    while (!regex_parse_done(p) && regex_parse_peek(p) != '>') {
        uint32_t point;
        bool valid = true, unicode_escape = false;
        if (regex_parse_take(p, '\\')) {
            if (!regex_parse_take(p, 'u')) valid = false;
            else if ((unicode_escape = true) && regex_parse_take(p, '{')) valid = regex_parse_hex_braced(p, &point);
            else {
                valid = regex_parse_hex_fixed(p, 4, &point);
                if (valid && point >= 0xd800 && point <= 0xdbff && regex_parse_match(p, "\\u")) {
                    size_t second_start = p->position;
                    uint32_t second;
                    p->position += 2;
                    if (regex_parse_hex_fixed(p, 4, &second) && second >= 0xdc00 && second <= 0xdfff)
                        point = 0x10000 + (point - 0xd800) * 0x400 + second - 0xdc00;
                    else p->position = second_start;
                }
            }
        } else {
            size_t width;
            point = regex_parse_rune(p, &width);
            valid = !(point == 0xfffd && width == 1);
            p->position += width;
        }
        if (!valid || !regex_parse_identifier(point, count == 0)) {
            regex_parse_fail(p, start, "invalid capture name", !valid && unicode_escape ? "Invalid Unicode escape" : "Invalid capture group name");
            return NULL;
        }
        count += regex_parse_encode(point, name + count);
    }
    if (count == 0 || !regex_parse_take(p, '>')) {
        regex_parse_fail(p, p->position, "invalid capture name", "Invalid capture group name");
        return NULL;
    }
    return name;
}
static regex_node *regex_parse_octal(regex_parser *p, size_t start) {
    unsigned char byte = regex_parse_peek(p);
    if (byte == '8' || byte == '9') { p->position++; return regex_parse_character(p, byte, 1, start); }
    unsigned limit = byte >= '4' ? 2 : 3;
    uint32_t value = 0;
    for (unsigned i = 0; i < limit && !regex_parse_done(p) && regex_parse_peek(p) >= '0' && regex_parse_peek(p) <= '7'; i++) {
        value = value * 8 + regex_parse_peek(p) - '0';
        p->position++;
    }
    return regex_parse_character(p, value, 1, start);
}
static regex_node *regex_parse_escape(regex_parser *p, bool in_class, bool *quantifiable) {
    size_t start = p->position;
    p->position++;
    *quantifiable = true;
    if (regex_parse_done(p)) return regex_parse_fail(p, p->position, "trailing escape", "\\ at end of pattern");
    if (regex_parse_peek(p) >= 0x80) {
        if (regex_parse_unicode(p)) return regex_parse_fail(p, start, "invalid identity escape", "Invalid escape");
        regex_node *node = regex_parse_literal(p);
        if (node != NULL) { node->kind = 1; node->start = start; }
        return node;
    }
    unsigned char byte = regex_parse_peek(p);
    p->position++;
    if (!in_class && (byte == 'b' || byte == 'B')) {
        regex_node *node = regex_parse_new(p, REGEX_PARSE_ASSERTION);
        if (node != NULL) node->kind = byte == 'b' ? 2 : 3;
        *quantifiable = false;
        return node;
    }
    if (byte >= '1' && byte <= '9') {
        p->position--;
        const char *digits = regex_parse_decimal(p);
        if (digits == NULL) return NULL;
        size_t number = 0;
        bool fits = true;
        for (const char *at = digits; *at != 0; at++) {
            unsigned digit = (unsigned)(*at - '0');
            if (number > (SIZE_MAX - digit) / 10) { fits = false; break; }
            number = number * 10 + digit;
        }
        if (!in_class && fits && number <= p->capture_count) {
            regex_node *node = regex_parse_new(p, REGEX_PARSE_REFERENCE);
            if (node != NULL) node->index = number;
            return node;
        }
        if (regex_parse_unicode(p)) return regex_parse_fail(p, start, "invalid decimal escape", in_class && byte <= '7' ? "Invalid decimal escape" : "Invalid escape");
        p->position = start + 1;
        return regex_parse_octal(p, start);
    }
    if (!in_class && byte == 'k') {
        if (!p->named_capture && !regex_parse_unicode(p)) return regex_parse_character(p, 'k', 1, start);
        if (!regex_parse_take(p, '<')) return regex_parse_fail(p, p->position, "invalid named reference", "Invalid named reference");
        const char *name = regex_parse_group_name(p);
        if (name == NULL) return NULL;
        regex_node *node = regex_parse_new(p, REGEX_PARSE_REFERENCE);
        if (node != NULL) { node->name = name; node->start = start; }
        return node;
    }
    if (byte == 'p' || byte == 'P') {
        if (!regex_parse_unicode(p)) return regex_parse_character(p, byte, 1, start);
        if (!regex_parse_take(p, '{')) return regex_parse_fail(p, start, "invalid property escape", in_class ? "Invalid property name in character class" : "Invalid property name");
        size_t property_start = p->position;
        while (p->position < p->result->length && regex_parse_peek(p) != '}') p->position++;
        if (p->position == p->result->length) return regex_parse_fail(p, property_start, "unterminated property escape", in_class ? "Invalid property name in character class" : "Invalid property name");
        const adamic_regex_compile_property *property = adamic_regex_compile_lookup_property(p->result->source + property_start, p->position - property_start, regex_parse_sets(p));
        p->position++;
        if (property == NULL || (byte == 'P' && property->string_property)) return regex_parse_fail(p, start, "invalid Unicode property", in_class ? "Invalid property name in character class" : "Invalid property name");
        regex_node *node = regex_parse_character(p, 0, 3, start);
        if (node != NULL) node->name = regex_parse_copy(p, property_start, p->position - 1);
        return node;
    }
    if (regex_parse_in(byte, "dDsSwW")) return regex_parse_character(p, 0, 2, start);
    if (in_class && byte == 'b') return regex_parse_character(p, 8, 1, start);
    if (byte == '0') {
        if (!regex_parse_done(p) && regex_parse_digit(regex_parse_peek(p))) {
            if (regex_parse_unicode(p)) return regex_parse_fail(p, start, "invalid decimal escape", "Invalid decimal escape");
            p->position = start + 1;
            return regex_parse_octal(p, start);
        }
        return regex_parse_character(p, 0, 1, start);
    }
    if (byte == 'c') {
        unsigned char next = regex_parse_peek(p);
        if (!regex_parse_done(p) && (regex_parse_in(next, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") ||
            (in_class && !regex_parse_unicode(p) && (regex_parse_digit(next) || next == '_')))) {
            p->position++;
            return regex_parse_character(p, next % 32, 1, start);
        }
        if (regex_parse_unicode(p)) return regex_parse_fail(p, start, "invalid control escape", "Invalid Unicode escape");
        p->position = start + 1;
        return regex_parse_character(p, '\\', 0, start);
    }
    if (byte == 'u') return regex_parse_unicode_escape(p, start);
    if (byte == 'x') {
        uint32_t point;
        if (!regex_parse_hex_fixed(p, 2, &point)) {
            if (regex_parse_unicode(p)) return regex_parse_fail(p, start, "invalid hex escape", "Invalid escape");
            p->position = start + 2;
            point = 'x';
        }
        return regex_parse_character(p, point, 1, start);
    }
    if (regex_parse_in(byte, "fnrtv")) {
        uint32_t value = byte == 'f' ? 12 : byte == 'n' ? 10 : byte == 'r' ? 13 : byte == 't' ? 9 : 11;
        return regex_parse_character(p, value, 1, start);
    }
    if (in_class && byte >= '1' && byte <= '7' && regex_parse_unicode(p)) return regex_parse_fail(p, start, "invalid identity escape", "Invalid decimal escape");
    if (regex_parse_unicode(p) && !regex_parse_in(byte, "^$\\.*+?()[]{}|/") &&
        !(in_class && (byte == '-' || (regex_parse_sets(p) && regex_parse_in(byte, "!#%&,:;<=>@`~")))))
        return regex_parse_fail(p, start, "invalid identity escape", in_class ? "Invalid escape" : "Invalid escape");
    return regex_parse_character(p, byte, 1, start);
}

static regex_node *regex_parse_disjunction(regex_parser *p, unsigned char stop);
static regex_node *regex_parse_class(regex_parser *p);
static regex_node *regex_parse_class_character(regex_parser *p) {
    if (p->pending != NULL) return regex_parse_literal(p);
    if (regex_parse_done(p)) return regex_parse_fail(p, p->position, "unterminated character class", "Unterminated character class");
    if (regex_parse_peek(p) == '\\') {
        bool quantifiable;
        return regex_parse_escape(p, true, &quantifiable);
    }
    if (regex_parse_sets(p) && (regex_parse_in(regex_parse_peek(p), "(){}|/-") ||
        (p->position + 1 < p->result->length && p->result->source[p->position] == p->result->source[p->position + 1] &&
         regex_parse_in(regex_parse_peek(p), "!#$%&*+,.:;<=>?@^`~"))))
        return regex_parse_fail(p, p->position, "reserved character in Unicode set", regex_parse_in(regex_parse_peek(p), "(){}|/-") ? "Invalid character in character class" : "Invalid set operation in character class");
    return regex_parse_literal(p);
}
static void regex_parse_append(regex_node **head, regex_node ***tail, regex_node *node) {
    if (*head == NULL) *head = node;
    else **tail = node;
    *tail = &node->next;
}
static bool regex_parse_class_strings(regex_parser *p, regex_node *node) {
    if (node == NULL) return false;
    switch (node->type) {
        case REGEX_PARSE_CLASS_STRING:
            for (regex_node *a = node->children; a != NULL; a = a->next)
                if (a->children == NULL || a->children->next != NULL) return true;
            break;
        case REGEX_PARSE_CLASS_CHARACTER:
            if (node->left->kind == 3) {
                const char *name = node->left->name;
                const adamic_regex_compile_property *property = adamic_regex_compile_lookup_property((const unsigned char *)name, strlen(name), true);
                return property != NULL && property->string_property;
            }
            break;
        case REGEX_PARSE_UNION:
            for (regex_node *child = node->children; child != NULL; child = child->next)
                if (regex_parse_class_strings(p, child)) return true;
            break;
        case REGEX_PARSE_INTERSECTION: return regex_parse_class_strings(p, node->left) && regex_parse_class_strings(p, node->right);
        case REGEX_PARSE_SUBTRACTION: return regex_parse_class_strings(p, node->left);
        default: break;
    }
    return false;
}
static regex_node *regex_parse_class_operand(regex_parser *p) {
    if (regex_parse_sets(p) && regex_parse_peek(p) == '[') {
        regex_node *node = regex_parse_class(p);
        if (node == NULL) return NULL;
        if (!node->negated) return node->left;
        regex_node *negation = regex_parse_new(p, REGEX_PARSE_NEGATION);
        if (negation != NULL) negation->left = node->left;
        return negation;
    }
    if (regex_parse_sets(p) && regex_parse_match(p, "\\q{")) {
        p->position += 3;
        regex_node *node = regex_parse_new(p, REGEX_PARSE_CLASS_STRING);
        regex_node *alternative = regex_parse_new(p, REGEX_PARSE_ALTERNATIVE);
        if (node == NULL || alternative == NULL) return NULL;
        node->children = alternative;
        regex_node **tail = &alternative->children;
        while (!regex_parse_done(p) && regex_parse_peek(p) != '}') {
            if (regex_parse_take(p, '|')) {
                alternative->next = regex_parse_new(p, REGEX_PARSE_ALTERNATIVE);
                alternative = alternative->next;
                if (alternative == NULL) return NULL;
                tail = &alternative->children;
                continue;
            }
            regex_node *character = regex_parse_class_character(p);
            if (character == NULL) return NULL;
            *tail = character;
            tail = &character->next;
        }
        if (!regex_parse_take(p, '}')) return regex_parse_fail(p, p->position, "unterminated class string", "Unterminated class string disjunction");
        return node;
    }
    regex_node *character = regex_parse_class_character(p);
    if (character == NULL) return NULL;
    if (regex_parse_peek(p) == '-' && p->position + 1 < p->result->length && p->result->source[p->position + 1] != ']' &&
        !(regex_parse_sets(p) && regex_parse_match(p, "--"))) {
        p->position++;
        regex_node *right = regex_parse_class_character(p);
        if (right == NULL) return NULL;
        if (character->kind >= 2 || right->kind >= 2) {
            if (regex_parse_unicode(p)) return regex_parse_fail(p, p->position, "invalid character class range", "Invalid character class");
            regex_node *node = regex_parse_new(p, REGEX_PARSE_UNION);
            regex_node *a = regex_parse_new(p, REGEX_PARSE_CLASS_CHARACTER);
            regex_node *dash = regex_parse_new(p, REGEX_PARSE_CLASS_CHARACTER);
            regex_node *b = regex_parse_new(p, REGEX_PARSE_CLASS_CHARACTER);
            if (node == NULL || a == NULL || dash == NULL || b == NULL) return NULL;
            a->left = character;
            dash->left = regex_parse_character(p, '-', 0, p->position);
            if (dash->left == NULL) return NULL;
            b->left = right;
            node->children = a; a->next = dash; dash->next = b;
            return node;
        }
        if (character->value > right->value) return regex_parse_fail(p, p->position, "invalid character class range", "Range out of order in character class");
        regex_node *node = regex_parse_new(p, REGEX_PARSE_RANGE);
        if (node != NULL) { node->left = character; node->right = right; }
        return node;
    }
    regex_node *node = regex_parse_new(p, REGEX_PARSE_CLASS_CHARACTER);
    if (node != NULL) node->left = character;
    return node;
}
static regex_node *regex_parse_finish_class(regex_parser *p, bool negated, regex_node *expression) {
    if (!regex_parse_take(p, ']')) return regex_parse_fail(p, p->position, "unterminated character class", regex_parse_sets(p) && !regex_parse_done(p) ? "Invalid set operation in character class" : "Unterminated character class");
    if (negated && regex_parse_class_strings(p, expression)) return regex_parse_fail(p, p->position, "cannot negate a class containing strings", "Negated character class may contain strings");
    regex_node *node = regex_parse_new(p, REGEX_PARSE_CLASS);
    if (node != NULL) { node->negated = negated; node->left = expression; }
    return node;
}
static regex_node *regex_parse_class(regex_parser *p) {
    p->position++;
    bool negated = regex_parse_take(p, '^');
    regex_node *node = regex_parse_new(p, REGEX_PARSE_UNION);
    if (node == NULL) return NULL;
    regex_node **tail = &node->children;
    size_t count = 0;
    while (!regex_parse_done(p) && regex_parse_peek(p) != ']') {
        regex_node *operand = regex_parse_class_operand(p);
        if (operand == NULL) return NULL;
        regex_parse_append(&node->children, &tail, operand);
        count++;
        if (regex_parse_sets(p) && (regex_parse_match(p, "&&") || regex_parse_match(p, "--"))) {
            if (count != 1) return regex_parse_fail(p, p->position, "invalid Unicode set operation", "Invalid set operation in character class");
            if (operand->type == REGEX_PARSE_RANGE) return regex_parse_fail(p, p->position, "range must be nested in Unicode set operation", "Invalid set operation in character class");
            bool intersection = regex_parse_peek(p) == '&';
            const char *operation = intersection ? "&&" : "--";
            p->position += 2;
            regex_node *left = node;
            for (;;) {
                regex_node *right = regex_parse_class_operand(p);
                if (right == NULL) return NULL;
                if (right->type == REGEX_PARSE_RANGE) return regex_parse_fail(p, p->position, "range must be nested in Unicode set operation", "Invalid set operation in character class");
                regex_node *combined = regex_parse_new(p, intersection ? REGEX_PARSE_INTERSECTION : REGEX_PARSE_SUBTRACTION);
                if (combined == NULL) return NULL;
                combined->left = left; combined->right = right; left = combined;
                if (!regex_parse_match(p, operation)) break;
                p->position += 2;
            }
            return regex_parse_finish_class(p, negated, left);
        }
    }
    return regex_parse_finish_class(p, negated, node);
}
static regex_node *regex_parse_group(regex_parser *p, bool *quantifiable) {
    p->position++;
    regex_node *node = regex_parse_new(p, REGEX_PARSE_GROUP);
    if (node == NULL) return NULL;
    if (regex_parse_take(p, '?')) {
        if (regex_parse_take(p, ':')) node->kind = 1;
        else if (regex_parse_take(p, '=')) node->kind = 2;
        else if (regex_parse_take(p, '!')) node->kind = 3;
        else if (regex_parse_take(p, '<')) {
            if (regex_parse_take(p, '=')) node->kind = 4;
            else if (regex_parse_take(p, '!')) node->kind = 5;
            else { node->name = regex_parse_group_name(p); if (node->name == NULL) return NULL; }
        } else if (regex_parse_in(regex_parse_peek(p), "ims-")) {
            bool disabling = false, any = false;
            unsigned seen = 0;
            while (!regex_parse_done(p) && regex_parse_peek(p) != ':') {
                unsigned char byte = regex_parse_peek(p);
                if (byte == '-' && !disabling) { disabling = true; p->position++; continue; }
                unsigned flag = byte == 'i' ? REGEX_FLAG_I : byte == 'm' ? REGEX_FLAG_M : byte == 's' ? REGEX_FLAG_S : 0;
                if (flag == 0 || (seen & flag) != 0) return regex_parse_fail(p, p->position, "invalid modifiers", (seen & flag) != 0 ? "Repeated flag in flag group" : "Invalid group");
                seen |= flag; any = true;
                if (disabling) node->disable |= flag; else node->enable |= flag;
                p->position++;
            }
            if (!any || !regex_parse_take(p, ':')) return regex_parse_fail(p, p->position, "invalid modifiers", "Invalid flag group");
            node->kind = 1;
        } else return regex_parse_fail(p, p->position, "invalid group", "Invalid group");
    }
    if (node->kind == 0) node->index = ++p->result->captures;
    node->left = regex_parse_disjunction(p, ')');
    if (node->left == NULL) return NULL;
    if (!regex_parse_take(p, ')')) return regex_parse_fail(p, p->position, "unterminated group", "Unterminated group");
    *quantifiable = node->kind < 2 || (!regex_parse_unicode(p) && node->kind < 4);
    return node;
}
static regex_node *regex_parse_term(regex_parser *p, bool *quantifiable) {
    *quantifiable = true;
    unsigned char byte = regex_parse_peek(p);
    if (byte == '^' || byte == '$') {
        p->position++;
        regex_node *node = regex_parse_new(p, REGEX_PARSE_ASSERTION);
        if (node != NULL) node->kind = byte == '^' ? 0 : 1;
        *quantifiable = false;
        return node;
    }
    if (byte == '.') { p->position++; return regex_parse_new(p, REGEX_PARSE_DOT); }
    if (byte == '[') return regex_parse_class(p);
    if (byte == '(') return regex_parse_group(p, quantifiable);
    if (byte == '\\') return regex_parse_escape(p, false, quantifiable);
    if (regex_parse_in(byte, "*+?")) return regex_parse_fail(p, p->position, "nothing to repeat", "Nothing to repeat");
    if (byte == '{') {
        if (regex_parse_quantifier_ahead(p)) return regex_parse_fail(p, p->position, "nothing to repeat", "Nothing to repeat");
        if (regex_parse_unicode(p)) return regex_parse_fail(p, p->position, "incomplete quantifier", "Lone quantifier brackets");
    }
    if ((byte == ']' || byte == '}') && regex_parse_unicode(p)) return regex_parse_fail(p, p->position, "lone punctuation", "Lone quantifier brackets");
    return regex_parse_literal(p);
}
static regex_node *regex_parse_quantifier(regex_parser *p, regex_node *atom) {
    regex_node *node = regex_parse_new(p, REGEX_PARSE_QUANTIFIER);
    if (node == NULL) return NULL;
    node->left = atom; node->minimum = "0";
    unsigned char byte = regex_parse_peek(p);
    p->position++;
    if (byte == '+') node->minimum = "1";
    else if (byte == '?') node->maximum = "1";
    else if (byte == '{') {
        size_t start = p->position - 1;
        node->minimum = regex_parse_decimal(p);
        if (node->minimum == NULL) return NULL;
        node->maximum = node->minimum;
        if (regex_parse_take(p, ',')) node->maximum = regex_parse_decimal(p);
        if (p->result->status != 0) return NULL;
        if (!regex_parse_take(p, '}')) return regex_parse_fail(p, start, "incomplete quantifier", "Incomplete quantifier");
        if (node->maximum != NULL && regex_parse_decimal_compare(node->minimum, node->maximum) > 0) {
            if (regex_parse_decimal_compare(node->maximum, "2147483647") >= 0) {
                p->result->status = 3;
                p->result->error_offset = start;
                p->result->reference_reason = "clamps quantifier bounds above 2^31-1 before the min > max check";
                p->result->message = "RegExp refused: V8 clamps quantifier bounds above 2^31-1 before the min > max check departs from ECMA-262 22.2.1.1 Static Semantics: Early Errors, QuantifierPrefix; native and JavaScript must agree";
                p->result->message_length = strlen(p->result->message);
                return NULL;
            }
            return regex_parse_fail(p, start, "quantifier range out of order", "numbers out of order in {} quantifier");
        }
    }
    node->greedy = !regex_parse_take(p, '?');
    return node;
}
static regex_node *regex_parse_disjunction(regex_parser *p, unsigned char stop) {
    regex_node *node = regex_parse_new(p, REGEX_PARSE_DISJUNCTION);
    if (node == NULL) return NULL;
    regex_node **alternatives_tail = &node->children;
    for (;;) {
        regex_node *alternative = regex_parse_new(p, REGEX_PARSE_ALTERNATIVE);
        if (alternative == NULL) return NULL;
        *alternatives_tail = alternative; alternatives_tail = &alternative->next;
        regex_node **tail = &alternative->children;
        while (!regex_parse_done(p) && regex_parse_peek(p) != '|' && regex_parse_peek(p) != stop) {
            if (regex_parse_peek(p) == ')' && stop == 0) return regex_parse_fail(p, p->position, "unmatched closing parenthesis", "Unmatched ')'");
            bool quantifiable;
            regex_node *term = regex_parse_term(p, &quantifiable);
            if (term == NULL) return NULL;
            if (regex_parse_unicode(p) && regex_parse_peek(p) == '{' && !regex_parse_quantifier_ahead(p)) return regex_parse_fail(p, p->position, "incomplete quantifier", "Incomplete quantifier");
            if (regex_parse_quantifier_ahead(p)) {
                if (term->type == REGEX_PARSE_ASSERTION) return regex_parse_fail(p, p->position, "nothing to repeat", "Nothing to repeat");
                size_t quantifier_start = p->position;
                term = regex_parse_quantifier(p, term);
                if (term == NULL) return NULL;
                if (!quantifiable) return regex_parse_fail(p, quantifier_start, "nothing to repeat", "Invalid quantifier");
                if (regex_parse_quantifier_ahead(p)) return regex_parse_fail(p, p->position, "nothing to repeat", "Nothing to repeat");
            }
            *tail = term; tail = &term->next;
        }
        if (regex_parse_done(p) || regex_parse_peek(p) != '|') break;
        p->position++;
    }
    return node;
}
static bool regex_parse_find_name(regex_node *node, const char *name) {
    if (node == NULL) return false;
    if (node->type == REGEX_PARSE_GROUP && node->kind == 0 && node->name != NULL && strcmp(node->name, name) == 0) return true;
    if (regex_parse_find_name(node->left, name) || regex_parse_find_name(node->right, name)) return true;
    for (regex_node *child = node->children; child != NULL; child = child->next)
        if (regex_parse_find_name(child, name)) return true;
    return false;
}
static bool regex_parse_references(regex_parser *p, regex_node *node) {
    if (node == NULL) return true;
    if (node->type == REGEX_PARSE_REFERENCE && node->name != NULL && !regex_parse_find_name(p->result->body, node->name)) {
        regex_parse_fail(p, node->start, "unknown capture name", "Invalid named capture referenced");
        return false;
    }
    if (!regex_parse_references(p, node->left) || !regex_parse_references(p, node->right)) return false;
    for (regex_node *child = node->children; child != NULL; child = child->next)
        if (!regex_parse_references(p, child)) return false;
    return true;
}
/* For a pair of names, simultaneous occurrence is possible unless some
 * disjunction places them in different alternatives. This is the Go path-map
 * rule, computed without retaining a path allocation per capture. */
static unsigned regex_parse_name_pair(regex_node *node, const regex_node *a, const regex_node *b, bool *disjoint) {
    if (node == NULL) return 0;
    unsigned found = node == a ? 1 : node == b ? 2 : 0;
    if (node->type == REGEX_PARSE_DISJUNCTION) {
        unsigned previous = 0;
        for (regex_node *child = node->children; child != NULL; child = child->next) {
            unsigned here = regex_parse_name_pair(child, a, b, disjoint);
            if ((here == 1 && (previous & 2)) || (here == 2 && (previous & 1))) *disjoint = true;
            previous |= here;
        }
        return found | previous;
    }
    found |= regex_parse_name_pair(node->left, a, b, disjoint) | regex_parse_name_pair(node->right, a, b, disjoint);
    for (regex_node *child = node->children; child != NULL; child = child->next)
        found |= regex_parse_name_pair(child, a, b, disjoint);
    return found;
}
static bool regex_parse_duplicate_visit(regex_parser *p, regex_node *node, regex_node *root, regex_node *target) {
    if (node == NULL) return true;
    if (node->type == REGEX_PARSE_GROUP && node->kind == 0 && node->name != NULL) {
        if (target == NULL) {
            if (!regex_parse_duplicate_visit(p, root, root, node)) return false;
        } else if (node != target && strcmp(node->name, target->name) == 0) {
            bool disjoint = false;
            regex_parse_name_pair(root, node, target, &disjoint);
            if (!disjoint) {
                regex_parse_fail(p, 0, "duplicate capture name", "Duplicate capture group name");
                return false;
            }
        }
    }
    if (!regex_parse_duplicate_visit(p, node->left, root, target) || !regex_parse_duplicate_visit(p, node->right, root, target)) return false;
    for (regex_node *child = node->children; child != NULL; child = child->next)
        if (!regex_parse_duplicate_visit(p, child, root, target)) return false;
    return true;
}
static bool regex_parse_flags(regex_parser *p, const unsigned char *flags, size_t length) {
    bool seen[128] = {false}, unicode = false, sets = false;
    for (size_t i = 0; i < length; i++) {
        unsigned char byte = flags[i];
        if (byte >= 128 || seen[byte]) { regex_parse_fail(p, i, "invalid or duplicate flag", "Invalid flags supplied to RegExp constructor"); return false; }
        seen[byte] = true;
        unsigned flag = 0;
        switch (byte) {
            case 'd': flag = REGEX_FLAG_D; break;
            case 'g': flag = REGEX_FLAG_G; break;
            case 'i': flag = REGEX_FLAG_I; break;
            case 'm': flag = REGEX_FLAG_M; break;
            case 's': flag = REGEX_FLAG_S; break;
            case 'u': flag = REGEX_FLAG_U; unicode = true; break;
            case 'v': flag = REGEX_FLAG_U | REGEX_FLAG_V; sets = true; break;
            case 'y': flag = REGEX_FLAG_Y; break;
            default: regex_parse_fail(p, i, "invalid flag", "Invalid flags supplied to RegExp constructor"); return false;
        }
        p->result->flags |= flag;
    }
    if (unicode && sets) { regex_parse_fail(p, 0, "u and v flags are mutually exclusive", "Invalid flags supplied to RegExp constructor"); return false; }
    return true;
}
static void regex_parse_pattern(const unsigned char *pattern, size_t length,
    const unsigned char *flags, size_t flag_length, adamic_regex_parse_result *result) {
    memset(result, 0, sizeof(*result));
    regex_parser parser = {.result = result};
    if (!regex_parse_flags(&parser, flags, flag_length)) return;
    if (length == SIZE_MAX) { result->status = 2; return; }
    unsigned char *copy = regex_parse_allocate(&parser, length + 1);
    if (copy == NULL) return;
    if (length != 0) memcpy(copy, pattern, length);
    result->source = copy; result->length = length;
    bool escaped = false, in_class = false;
    for (size_t i = 0; i < length; i++) {
        unsigned char byte = pattern[i];
        if (escaped) { escaped = false; continue; }
        if (byte == '\\') { escaped = true; continue; }
        if (byte == '[') { in_class = true; continue; }
        if (byte == ']') { in_class = false; continue; }
        if (!in_class && byte == '(' && (i + 1 >= length || pattern[i + 1] != '?' ||
            (i + 2 < length && pattern[i + 1] == '?' && pattern[i + 2] == '<' &&
             (i + 3 >= length || (pattern[i + 3] != '=' && pattern[i + 3] != '!'))))) {
            parser.capture_count++;
            if (i + 2 < length && pattern[i + 1] == '?' && pattern[i + 2] == '<') parser.named_capture = true;
        }
    }
    result->body = regex_parse_disjunction(&parser, 0);
    if (result->body == NULL) return;
    if (!regex_parse_done(&parser)) { regex_parse_fail(&parser, parser.position, "unexpected character", "Invalid escape"); return; }
    if (!regex_parse_references(&parser, result->body)) return;
    regex_parse_duplicate_visit(&parser, result->body, result->body, NULL);
}
void adamic_regex_parse(const unsigned char *pattern, size_t length,
    const unsigned char *flags, size_t flag_length, adamic_regex_parse_result *result) {
    regex_parse_pattern(pattern, length, flags, flag_length, result);
    if (result->status != 1) return;
    bool flag_error = strcmp(result->node_reason, "Invalid flags supplied to RegExp constructor") == 0;
    const char *prefix = flag_error ? "Invalid flags supplied to RegExp constructor '" : "Invalid regular expression: /";
    size_t prefix_length = strlen(prefix), reason_length = flag_error ? 0 : strlen(result->node_reason);
    size_t pattern_length = flag_error ? 0 : length;
    size_t overhead = prefix_length + reason_length + (flag_error ? 1 : 3);
    if (pattern_length >= SIZE_MAX - overhead || flag_length > SIZE_MAX - overhead - pattern_length - 1) {
        result->status = 2; return;
    }
    size_t total = overhead + pattern_length + flag_length;
    regex_parser parser = {.result = result};
    char *message = regex_parse_allocate(&parser, total + 1);
    if (message == NULL) return;
    size_t position = 0;
    memcpy(message, prefix, prefix_length); position += prefix_length;
    if (!flag_error) {
        if (length != 0) memcpy(message + position, pattern, length);
        position += length; message[position++] = '/';
    }
    if (flag_length != 0) memcpy(message + position, flags, flag_length);
    position += flag_length;
    if (flag_error) message[position++] = '\'';
    else {
        message[position++] = ':'; message[position++] = ' ';
        memcpy(message + position, result->node_reason, reason_length); position += reason_length;
    }
    message[position] = 0;
    result->message = message; result->message_length = position;
}
#endif
