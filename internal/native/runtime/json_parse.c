// JSON grammar, scanning, continuation stack and SyntaxError diagnostics ported from V8
// 13.6.233.17 src/json/json-parser.cc, json-parser.h and src/common/message-template.h.
// Copyright the V8 project authors. BSD-3-Clause; see THIRD_PARTY_NOTICES.md.
// Numeric conversion is the existing V8 port in parse.c. Transient parse nodes have explicit,
// iterative cleanup; they are not a garbage collector or an any-valued language representation.
#include "json_parse.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct json_node json_node;
typedef struct json_member { adamic_string *key; json_node *value; } json_member;
struct json_node {
 enum adamic_json_kind kind;
 adamic_value value;
 size_t count, capacity;
 json_member *members;
 json_node *allocated_next;
 adamic_json_schema schema;
 adamic_json_field *fields;
 adamic_shape shape;
 const char **names;
 bool *references;
};
typedef struct json_frame { json_node *node; adamic_string *key; } json_frame;
typedef struct json_parser {
 const adamic_string *text;
 size_t position, length;
 json_node *allocated;
 json_frame *frames;
 size_t depth, capacity;
} json_parser;

static void *json_grow(void *memory, size_t count, size_t size) {
 if (count > SIZE_MAX / size) {
  static const char message[] = "out of memory";
  adamic_panic(message, sizeof message - 1);
 }
 void *grown = realloc(memory, count * size);
 if (grown == NULL) {
  static const char message[] = "out of memory";
  adamic_panic(message, sizeof message - 1);
 }
 return grown;
}
static unsigned character(const json_parser *p) {
 return p->position == p->length ? UINT32_MAX : (unsigned)adamic_string_char_code_at(p->text, (double)p->position);
}
static void whitespace(json_parser *p) {
 while (p->position < p->length) {
  unsigned c = character(p);
  if (c != ' ' && c != '\t' && c != '\r' && c != '\n') break;
  p->position++;
 }
}
static adamic_string *ascii_text(const char *text) {
 adamic_string part = ADAMIC_STRING_BYTES(text, strlen(text));
 return adamic_string_concat(1, (adamic_string *const[]){&part});
}
static void syntax_error(json_parser *p, const char *prefix) {
 if (adamic_thrown != NULL) return;
 adamic_string *message;
 unsigned c = character(p);
 if (prefix == NULL && c == UINT32_MAX) {
  message = ascii_text("Unexpected end of JSON input");
 } else if (prefix != NULL || c == '"' || c == '-' || (c >= '0' && c <= '9')) {
  if (prefix == NULL) prefix = c == '"' ? "Unexpected string in JSON" : "Unexpected number in JSON";
  size_t line = 1, last = 0;
  for (size_t index = 0; index < p->position; index++) {
   unsigned unit = (unsigned)adamic_string_char_code_at(p->text, (double)index);
   if (unit == '\r' && index + 1 < p->position && adamic_string_char_code_at(p->text, (double)(index + 1)) == '\n') {
    index++; unit = '\n';
   }
   if (unit == '\r' || unit == '\n') { line++; last = index + 1; }
  }
  char buffer[256];
  (void)snprintf(buffer, sizeof buffer, "%s at position %zu (line %zu column %zu)", prefix, p->position, line, 1 + p->position - last);
  message = ascii_text(buffer);
 } else {
  static const char *const special[] = {"[object Object]", "undefined", "Infinity", "NaN"};
  bool short_message = false;
  for (size_t index = 0; index < sizeof special / sizeof special[0]; index++) {
   size_t length = strlen(special[index]);
   if (p->text->length == length && memcmp(p->text->bytes, special[index], length) == 0) { short_message = true; break; }
  }
  adamic_string *parts[5];
  size_t count = 0;
  if (short_message) {
   parts[count++] = ascii_text("\"");
   parts[count++] = adamic_retain((adamic_string *)p->text);
   parts[count++] = ascii_text("\" is not valid JSON");
  } else {
   size_t start = 0, end = p->length;
   bool before = false, after = false;
   if (p->length >= 21) {
    if (p->position < 10) { end = p->position + 10; after = true; }
    else if (p->position < p->length - 10) { start = p->position - 10; end = p->position + 10; before = after = true; }
    else { start = p->position - 10; before = true; }
   }
   parts[count++] = ascii_text("Unexpected token '");
   parts[count++] = adamic_string_slice(p->text, (double)p->position, (double)(p->position + 1), true);
   parts[count++] = ascii_text(before ? "', ...\"" : "', \"");
   parts[count++] = adamic_string_slice(p->text, (double)start, (double)end, true);
   parts[count++] = ascii_text(after ? "\"... is not valid JSON" : "\" is not valid JSON");
  }
  message = adamic_string_concat(count, parts);
  for (size_t index = 0; index < count; index++) adamic_release(parts[index]);
 }
 adamic_thrown = adamic_builtin_error_new(2, message);
 adamic_release(message);
}
static int hex_digit(unsigned c) {
 if (c >= '0' && c <= '9') return (int)(c - '0');
 if (c >= 'a' && c <= 'f') return (int)(c - 'a' + 10);
 if (c >= 'A' && c <= 'F') return (int)(c - 'A' + 10);
 return -1;
}
static adamic_string *scan_string(json_parser *p) {
 p->position++; // Opening quotation mark.
 double *units = NULL;
 size_t count = 0, capacity = 0;
 adamic_string *result = NULL;
 while (true) {
  unsigned c = character(p);
  if (c == UINT32_MAX) { syntax_error(p, "Unterminated string in JSON"); break; }
  if (c == '"') { p->position++; result = adamic_string_from_char_codes(count, units); break; }
  if (c == '\\') {
   p->position++; c = character(p);
   if (c > 255) { syntax_error(p, NULL); break; }
   switch (c) {
   case '"': case '\\': case '/': break;
   case 'b': c = '\b'; break;
   case 't': c = '\t'; break;
   case 'n': c = '\n'; break;
   case 'f': c = '\f'; break;
   case 'r': c = '\r'; break;
   case 'u': {
    unsigned decoded = 0;
    for (size_t index = 0; index < 4; index++) {
     if (p->position < p->length) p->position++;
     int digit = hex_digit(character(p));
     if (digit < 0) { syntax_error(p, "Bad Unicode escape in JSON"); break; }
     decoded = decoded * 16 + (unsigned)digit;
    }
    c = decoded;
    break;
   }
   default: syntax_error(p, "Bad escaped character in JSON"); break;
   }
   if (adamic_thrown != NULL) break;
  } else if (c < 0x20) { syntax_error(p, "Bad control character in string literal in JSON"); break; }
  if (count == capacity) { capacity = capacity == 0 ? 16 : capacity * 2; units = json_grow(units, capacity, sizeof *units); }
  units[count++] = (double)c;
  p->position++;
 }
 free(units);
 return result;
}
static bool decimal(unsigned c) { return c >= '0' && c <= '9'; }
static double scan_number(json_parser *p) {
 size_t start = p->position;
 if (character(p) == '-') p->position++;
 if (character(p) == '0') {
  p->position++;
  if (decimal(character(p))) { syntax_error(p, "Unexpected number in JSON"); return 0; }
 } else {
  if (!decimal(character(p))) { syntax_error(p, "No number after minus sign in JSON"); return 0; }
  while (decimal(character(p))) p->position++;
 }
 if (character(p) == '.') {
  p->position++;
  if (!decimal(character(p))) { syntax_error(p, "Unterminated fractional number in JSON"); return 0; }
  while (decimal(character(p))) p->position++;
 }
 if (character(p) == 'e' || character(p) == 'E') {
  p->position++;
  if (character(p) == '+' || character(p) == '-') p->position++;
  if (!decimal(character(p))) { syntax_error(p, "Exponent part is missing a number in JSON"); return 0; }
  while (decimal(character(p))) p->position++;
 }
 adamic_string *token = adamic_string_slice(p->text, (double)start, (double)p->position, true);
 double value = adamic_number_parse_float(token);
 adamic_release(token);
 return value;
}
static void scan_literal(json_parser *p, const char *literal) {
 for (size_t index = 0; literal[index] != 0; index++) {
  if (character(p) != (unsigned char)literal[index]) { syntax_error(p, NULL); return; }
  p->position++;
 }
}
static json_node *make_node(json_parser *p, enum adamic_json_kind kind) {
 json_node *node = json_grow(NULL, 1, sizeof *node);
 memset(node, 0, sizeof *node);
 node->kind = kind;
 node->allocated_next = p->allocated;
 p->allocated = node;
 return node;
}
static void push(json_parser *p, json_node *node) {
 if (p->depth == p->capacity) { p->capacity = p->capacity == 0 ? 16 : p->capacity * 2; p->frames = json_grow(p->frames, p->capacity, sizeof *p->frames); }
 p->frames[p->depth++] = (json_frame){node, NULL};
}
static bool property(json_parser *p, json_frame *frame, bool first) {
 whitespace(p);
 if (character(p) != '"') {
  syntax_error(p, first ? "Expected property name or '}' in JSON" : "Expected double-quoted property name in JSON");
  return false;
 }
 frame->key = scan_string(p);
 if (adamic_thrown != NULL) return false;
 whitespace(p);
 if (character(p) != ':') { syntax_error(p, "Expected ':' after property name in JSON"); return false; }
 p->position++;
 return true;
}
static void add_member(json_node *parent, adamic_string *key, json_node *value) {
 if (parent->kind == adamic_json_object) {
  for (size_t index = 0; index < parent->count; index++) {
   if (adamic_string_equal(parent->members[index].key, key)) {
    // CreateDataProperty replaces the value without moving the key's insertion position.
    parent->members[index].value = value;
    adamic_release(key);
    return;
   }
  }
 }
 if (parent->count == parent->capacity) { parent->capacity = parent->capacity == 0 ? 8 : parent->capacity * 2; parent->members = json_grow(parent->members, parent->capacity, sizeof *parent->members); }
 parent->members[parent->count++] = (json_member){key, value};
}
// The continuation stack is V8's ParseJsonValue, with explicitly owned transient values.
static json_node *parse_value(json_parser *p) {
 json_node *value;
produce:
 whitespace(p);
 unsigned c = character(p);
 if (c == '{' || c == '[') {
  bool object = c == '{';
  value = make_node(p, object ? adamic_json_object : adamic_json_tuple);
  p->position++; whitespace(p);
  if (character(p) == (object ? '}' : ']')) { p->position++; goto consume; }
  push(p, value);
  if (object && !property(p, &p->frames[p->depth - 1], true)) return NULL;
  goto produce;
 }
 if (c == '"') { value = make_node(p, adamic_json_string); value->value.reference = scan_string(p); }
 else if (c == '-' || decimal(c)) { value = make_node(p, adamic_json_number); value->value.number = scan_number(p); }
 else if (c == 't' || c == 'f') { value = make_node(p, adamic_json_boolean); value->value.boolean = c == 't'; scan_literal(p, c == 't' ? "true" : "false"); }
 else if (c == 'n') { value = make_node(p, adamic_json_null); scan_literal(p, "null"); }
 else { syntax_error(p, NULL); return NULL; }
 if (adamic_thrown != NULL) return NULL;
consume:
 if (p->depth == 0) {
  whitespace(p);
  if (p->position != p->length) { syntax_error(p, "Unexpected non-whitespace character after JSON"); return NULL; }
  return value;
 }
 json_frame *frame = &p->frames[p->depth - 1];
 json_node *parent = frame->node;
 add_member(parent, frame->key, value); frame->key = NULL;
 whitespace(p);
 if (character(p) == ',') {
  p->position++;
  if (parent->kind == adamic_json_object && !property(p, frame, false)) return NULL;
  goto produce;
 }
 if (character(p) != (parent->kind == adamic_json_object ? '}' : ']')) {
  syntax_error(p, parent->kind == adamic_json_object ? "Expected ',' or '}' after property value in JSON" : "Expected ',' or ']' after array element in JSON");
  return NULL;
 }
 p->position++; p->depth--; value = parent; goto consume;
}
static bool index_key(const adamic_string *key, uint32_t *result) {
 if (key->length == 0 || key->length > 10 || (key->length > 1 && key->bytes[0] == '0')) return false;
 uint64_t value = 0;
 for (size_t index = 0; index < key->length; index++) {
  unsigned char c = (unsigned char)key->bytes[index];
  if (c < '0' || c > '9') return false;
  value = value * 10 + c - '0';
 }
 if (value >= UINT32_MAX) return false;
 *result = (uint32_t)value;
 return true;
}
static void order_properties(json_parser *p) {
 for (json_node *node = p->allocated; node != NULL; node = node->allocated_next) {
  if (node->kind != adamic_json_object) continue;
  // Stable insertion order for other strings; integer indexes precede them numerically.
  for (size_t index = 1; index < node->count; index++) {
   json_member member = node->members[index];
   uint32_t number;
   if (!index_key(member.key, &number)) continue;
   size_t before = index;
   while (before > 0) {
    uint32_t previous;
    if (index_key(node->members[before - 1].key, &previous) && previous < number) break;
    node->members[before] = node->members[before - 1]; before--;
   }
   node->members[before] = member;
  }
 }
}
// Lowering bounds reviver input nesting to 64 and excludes access to holders/containers. The
// complete descendants are visited before the holder, in ECMAScript own-key order.
static void revive(json_node *node, adamic_string *key, adamic_closure *callback, enum adamic_json_kind returned, size_t takes) {
 for (size_t index = 0; index < node->count; index++) {
  json_member member = node->members[index];
  adamic_string *name = member.key == NULL ? adamic_string_from_number((double)index) : adamic_retain(member.key);
  revive(member.value, name, callback, returned, takes);
  adamic_release(name);
  if (adamic_thrown != NULL) return;
 }
 adamic_value arguments[2] = {{.reference = key}, node->value};
 (void)takes; // A callback with fewer parameters ignores the additional supplied slots.
 adamic_value result = callback->code(callback, arguments);
 if (adamic_thrown != NULL) return;
 if (node->kind == adamic_json_string) adamic_release(node->value.reference);
 node->kind = returned;
 node->value = result;
 if (returned == adamic_json_string && result.reference == NULL) node->kind = adamic_json_undefined;
}
// Only canonical literal parses reach this conversion, with a lowering-proven depth <= 64.
// Its ordinary objects and descriptors let the existing stringify implementation do all
// escaping and binary64 formatting, rather than adding a second number printer.
static adamic_value materialize(json_node *node) {
 node->schema = (adamic_json_schema){node->kind, NULL, 0, NULL, false};
 if (node->kind == adamic_json_string) return (adamic_value){.reference = adamic_retain(node->value.reference)};
 if (node->kind != adamic_json_object && node->kind != adamic_json_tuple) return node->value;
 node->fields = node->count == 0 ? NULL : json_grow(NULL, node->count, sizeof *node->fields);
 node->names = node->count == 0 ? NULL : json_grow(NULL, node->count, sizeof *node->names);
 node->references = node->count == 0 ? NULL : json_grow(NULL, node->count, sizeof *node->references);
 node->shape = (adamic_shape){node->count, node->names, node->references, NULL};
 // Shape names are private, since serialization reads slots; decoded JSON keys are in fields.
 for (size_t index = 0; index < node->count; index++) { node->names[index] = ""; node->references[index] = false; }
 adamic_object *object = adamic_object_new(&node->shape);
 for (size_t index = 0; index < node->count; index++) {
  json_node *child = node->members[index].value;
  object->slots[index] = materialize(child);
  node->references[index] = child->kind != adamic_json_number && child->kind != adamic_json_boolean;
  node->fields[index] = (adamic_json_field){node->members[index].key, index, &child->schema};
 }
 node->schema.count = node->count; node->schema.fields = node->fields;
 return (adamic_value){.reference = object};
}
static void cleanup(json_parser *p) {
 for (size_t index = 0; index < p->depth; index++) adamic_release(p->frames[index].key);
 free(p->frames);
 json_node *node = p->allocated;
 while (node != NULL) {
  json_node *next = node->allocated_next;
  if (node->kind == adamic_json_string) adamic_release(node->value.reference);
  for (size_t index = 0; index < node->count; index++) adamic_release(node->members[index].key);
  free(node->members); free(node->fields); free(node->names); free(node->references); free(node);
  node = next;
 }
}
adamic_value adamic_json_parse(const adamic_string *text, adamic_closure *reviver,
 enum adamic_json_kind returned, size_t takes, enum adamic_json_parse_mode mode) {
 static adamic_string undefined_text = ADAMIC_STRING("undefined");
 if (text == NULL) text = &undefined_text;
 json_parser parser = {text, 0, (size_t)adamic_string_length(text), NULL, NULL, 0, 0};
 json_node *root = parse_value(&parser);
 adamic_value result = {.number = 0};
 if (root == NULL) { cleanup(&parser); return result; }
 order_properties(&parser);
 if (reviver != NULL) {
  static adamic_string empty = ADAMIC_STRING("");
  revive(root, &empty, reviver, returned, takes);
  if (adamic_thrown != NULL) { cleanup(&parser); return result; }
 }
 switch (mode) {
 case adamic_json_parse_discard: break;
 case adamic_json_parse_number: case adamic_json_parse_boolean: result = root->value; break;
 case adamic_json_parse_string: result.reference = adamic_retain(root->value.reference); break;
 case adamic_json_parse_canonical: {
  adamic_value value = materialize(root);
  adamic_string *serialized = adamic_json_stringify(value, &root->schema, (adamic_value){.reference = NULL}, NULL, (adamic_value){.reference = NULL}, NULL);
  if (root->kind != adamic_json_number && root->kind != adamic_json_boolean) adamic_release(value.reference);
  static const char *const names[] = {"adamicJSONParsed"};
  static const bool references[] = {true};
  static const adamic_shape shape = {1, names, references, NULL};
  adamic_object *carrier = adamic_object_new(&shape);
  carrier->slots[0].reference = serialized;
  result.reference = carrier;
  break;
 }
 }
 cleanup(&parser);
 return result;
}
