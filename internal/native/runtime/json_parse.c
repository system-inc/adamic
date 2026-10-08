// Node 24.19.0 bundled V8 JSON grammar, diagnostics and context windows ported
// from deps/v8/src/json/json-parser.cc and src/common/message-template.h.
// The iterative container stack keeps valid deep JSON independent of C stack
// depth.
#include "json_parse.h"
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct json_node json_node;
struct json_node {
  enum json_parse_kind kind;
  adamic_value value;
  json_node **children;
  adamic_string **keys;
  size_t count, capacity;
  json_node *next;
};
typedef struct json_frame {
  json_node *node;
  int state;
  adamic_string *key;
} json_frame;
typedef struct json_parser {
  const adamic_string *source;
  double *units;
  size_t length, position;
  json_node *allocated;
  json_frame *stack;
  size_t depth, capacity;
  bool failed;
} json_parser;
static void *memory(size_t n) {
  void *p = calloc(1, n == 0 ? 1 : n);
  if (p == NULL) {
    adamic_panic("out of memory", 13);
  }
  return p;
}
static void *grow(void *p, size_t n) {
  void *r = realloc(p, n);
  if (r == NULL) {
    adamic_panic("out of memory", 13);
  }
  return r;
}
static int current(const json_parser *p) {
  return p->position == p->length ? -1 : (int)p->units[p->position];
}
static void whitespace(json_parser *p) {
  while (current(p) == ' ' || current(p) == '\t' || current(p) == '\r' ||
         current(p) == '\n') {
    p->position++;
  }
}
static adamic_string *text(const char *s) {
  return adamic_decode_utf8((const unsigned char *)s, strlen(s));
}
static adamic_string *slice(json_parser *p, size_t start, size_t end) {
  return adamic_string_from_char_codes(end - start, p->units + start);
}
static adamic_string *join(adamic_string *a, adamic_string *b) {
  adamic_string *r = adamic_string_concat(2, (adamic_string *const[]){a, b});
  adamic_release(a);
  adamic_release(b);
  return r;
}
static void fail(json_parser *p, const char *reason) {
  if (p->failed) {
    return;
  }
  p->failed = true;
  adamic_string *message;
  if (reason != NULL) {
    size_t line = 1, last = 0;
    for (size_t i = 0; i < p->position; i++) {
      if (p->units[i] == '\r' && i + 1 < p->position &&
          p->units[i + 1] == '\n') {
        i++;
      }
      if (p->units[i] == '\r' || p->units[i] == '\n') {
        line++;
        last = i + 1;
      }
    }
    char suffix[160];
    snprintf(suffix, sizeof suffix, "%s at position %zu (line %zu column %zu)",
             strcmp(reason, "Unexpected non-whitespace character after JSON") ==
                     0
                 ? ""
                 : " in JSON",
             p->position, line, p->position - last + 1);
    message = join(text(reason), text(suffix));
  } else if (current(p) < 0) {
    message = text("Unexpected end of JSON input");
  } else {
    static const char *const special[] = {"undefined", "NaN", "Infinity",
                                          "[object Object]"};
    bool is_special = false;
    for (size_t i = 0; i < 4; i++) {
      size_t n = strlen(special[i]);
      if (n != p->length) {
        continue;
      }
      bool same = true;
      for (size_t j = 0; j < n; j++) {
        if (p->units[j] != (unsigned char)special[i][j]) {
          same = false;
          break;
        }
      }
      if (same) {
        is_special = true;
        break;
      }
    }
    size_t start = 0, end = p->length;
    const char *prefix = "", *suffix = "";
    if (is_special) {
      message = join(join(text("\""), slice(p, 0, p->length)),
                     text("\" is not valid JSON"));
    } else {
      if (p->length >= 21) {
        if (p->position < 10) {
          end = p->position + 10;
          suffix = "...";
        } else if (p->position < p->length - 10) {
          start = p->position - 10;
          end = p->position + 10;
          prefix = "...";
          suffix = "...";
        } else {
          start = p->position - 10;
          prefix = "...";
        }
      }
      message = join(text("Unexpected token '"),
                     slice(p, p->position, p->position + 1));
      message = join(message, text("', "));
      message = join(message, text(prefix));
      message = join(message, text("\""));
      message = join(message, slice(p, start, end));
      message = join(message, text("\""));
      message = join(message, text(suffix));
      message = join(message, text(" is not valid JSON"));
    }
  }
  static adamic_string name = ADAMIC_STRING("SyntaxError");
  adamic_thrown = adamic_error_new(message);
  adamic_release(message);
  adamic_release(adamic_thrown->slots[0].reference);
  adamic_thrown->slots[0].reference = &name;
  adamic_error_tag(adamic_thrown);
}
static void unexpected(json_parser *p) {
  int c = current(p);
  if (c == '"') {
    fail(p, "Unexpected string");
  } else if (c == '-' || (c >= '0' && c <= '9')) {
    fail(p, "Unexpected number");
  } else {
    fail(p, NULL);
  }
}
static json_node *node(json_parser *p, enum json_parse_kind kind) {
  json_node *n = memory(sizeof *n);
  n->kind = kind;
  n->next = p->allocated;
  p->allocated = n;
  return n;
}
static int hex(int c) {
  if (c >= '0' && c <= '9') {
    return c - '0';
  }
  if (c >= 'a' && c <= 'f') {
    return c - 'a' + 10;
  }
  if (c >= 'A' && c <= 'F') {
    return c - 'A' + 10;
  }
  return -1;
}
static adamic_string *string(json_parser *p) {
  p->position++;
  size_t count = 0, capacity = 32;
  double *codes = memory(capacity * sizeof *codes);
  while (!p->failed) {
    int c = current(p);
    if (c < 0) {
      fail(p, "Unterminated string");
      break;
    }
    if (c == '"') {
      p->position++;
      adamic_string *r = adamic_string_from_char_codes(count, codes);
      free(codes);
      return r;
    }
    if (c < 32) {
      fail(p, "Bad control character in string literal");
      break;
    }
    if (c == '\\') {
      p->position++;
      c = current(p);
      switch (c) {
      case '"':
      case '\\':
      case '/':
        break;
      case 'b':
        c = '\b';
        break;
      case 'f':
        c = '\f';
        break;
      case 'n':
        c = '\n';
        break;
      case 'r':
        c = '\r';
        break;
      case 't':
        c = '\t';
        break;
      case 'u': {
        unsigned value = 0;
        for (int i = 0; i < 4; i++) {
          p->position++;
          int h = hex(current(p));
          if (h < 0) {
            fail(p, "Bad Unicode escape");
            break;
          }
          value = value * 16 + (unsigned)h;
        }
        c = (int)value;
        break;
      }
      default:
        if (c < 0 || c > 255) {
          unexpected(p);
        } else {
          fail(p, "Bad escaped character");
        }
        break;
      }
      if (p->failed) {
        break;
      }
    }
    if (count == capacity) {
      capacity *= 2;
      codes = grow(codes, capacity * sizeof *codes);
    }
    codes[count++] = c;
    p->position++;
  }
  free(codes);
  return NULL;
}
static bool digit(int c) { return c >= '0' && c <= '9'; }
static json_node *number(json_parser *p) {
  size_t start = p->position;
  if (current(p) == '-') {
    p->position++;
  }
  if (!digit(current(p))) {
    fail(p, "No number after minus sign");
    return NULL;
  }
  if (current(p) == '0') {
    p->position++;
    if (digit(current(p))) {
      unexpected(p);
      return NULL;
    }
  } else {
    while (digit(current(p))) {
      p->position++;
    }
  }
  if (current(p) == '.') {
    p->position++;
    if (!digit(current(p))) {
      fail(p, "Unterminated fractional number");
      return NULL;
    }
    while (digit(current(p))) {
      p->position++;
    }
  }
  if (current(p) == 'e' || current(p) == 'E') {
    p->position++;
    if (current(p) == '+' || current(p) == '-') {
      p->position++;
    }
    if (!digit(current(p))) {
      fail(p, "Exponent part is missing a number");
      return NULL;
    }
    while (digit(current(p))) {
      p->position++;
    }
  }
  size_t n = p->position - start;
  char *bytes = memory(n + 1);
  for (size_t i = 0; i < n; i++) {
    bytes[i] = (char)p->units[start + i];
  }
  json_node *r = node(p, json_number);
  r->value.number = strtod(bytes, NULL);
  free(bytes);
  return r;
}
static void push(json_parser *p, json_node *n) {
  if (p->depth == p->capacity) {
    p->capacity = p->capacity == 0 ? 32 : p->capacity * 2;
    p->stack = grow(p->stack, p->capacity * sizeof *p->stack);
  }
  p->stack[p->depth++] = (json_frame){n, 0, NULL};
}
static json_node *value(json_parser *p) {
  whitespace(p);
  int c = current(p);
  if (c == '{' || c == '[') {
    p->position++;
    json_node *r = node(p, c == '{' ? json_object : json_array);
    push(p, r);
    return r;
  }
  if (c == '"') {
    json_node *r = node(p, json_string);
    r->value.reference = string(p);
    return r;
  }
  if (c == '-' || digit(c)) {
    return number(p);
  }
  const char *word = c == 't'   ? "true"
                     : c == 'f' ? "false"
                     : c == 'n' ? "null"
                                : NULL;
  if (word != NULL) {
    for (size_t i = 0; word[i] != 0; i++) {
      if (current(p) != word[i]) {
        unexpected(p);
        return NULL;
      }
      p->position++;
    }
    json_node *r = node(p, c == 'n' ? json_null : json_boolean);
    r->value.boolean = c == 't';
    return r;
  }
  unexpected(p);
  return NULL;
}
static void add(json_node *n, adamic_string *key, json_node *child) {
  if (n->kind == json_object) {
    for (size_t i = 0; i < n->count; i++) {
      if (adamic_string_equal(n->keys[i], key)) {
        adamic_release(key);
        n->children[i] = child;
        return;
      }
    }
  }
  if (n->count == n->capacity) {
    n->capacity = n->capacity == 0 ? 8 : n->capacity * 2;
    n->children = grow(n->children, n->capacity * sizeof *n->children);
    if (n->kind == json_object) {
      n->keys = grow(n->keys, n->capacity * sizeof *n->keys);
    }
  }
  n->children[n->count] = child;
  if (n->kind == json_object) {
    n->keys[n->count] = key;
  }
  n->count++;
}
static json_node *parse(json_parser *p) {
  json_node *root = value(p);
  while (p->depth != 0 && !p->failed) {
    whitespace(p);
    size_t top = p->depth - 1;
    json_frame *f = &p->stack[top];
    json_node *n = f->node;
    int c = current(p);
    if (n->kind == json_array) {
      if (f->state == 0 || f->state == 2) {
        if (c == ']' && f->state == 0) {
          p->position++;
          p->depth--;
          continue;
        }
        f->state = 1;
        json_node *v = value(p);
        if (v != NULL) {
          add(n, NULL, v);
        }
        continue;
      }
      if (c == ']') {
        p->position++;
        p->depth--;
        continue;
      }
      if (c != ',') {
        fail(p, "Expected ',' or ']' after array element");
        break;
      }
      p->position++;
      f->state = 2;
      continue;
    }
    if (f->state == 0 || f->state == 3) {
      if (c == '}' && f->state == 0) {
        p->position++;
        p->depth--;
        continue;
      }
      if (c != '"') {
        fail(p, f->state == 0 ? "Expected property name or '}'"
                              : "Expected double-quoted property name");
        break;
      }
      f->key = string(p);
      if (p->failed) {
        break;
      }
      f->state = 1;
      continue;
    }
    if (f->state == 1) {
      if (c != ':') {
        fail(p, "Expected ':' after property name");
        break;
      }
      p->position++;
      adamic_string *key = f->key;
      f->key = NULL;
      f->state = 2;
      json_node *v = value(p);
      if (v != NULL) {
        add(n, key, v);
      } else {
        adamic_release(key);
      }
      continue;
    }
    if (c == '}') {
      p->position++;
      p->depth--;
      continue;
    }
    if (c != ',') {
      fail(p, "Expected ',' or '}' after property value");
      break;
    }
    p->position++;
    f->state = 3;
  }
  if (!p->failed) {
    whitespace(p);
    if (p->position != p->length) {
      fail(p, "Unexpected non-whitespace character after JSON");
    }
  }
  return root;
}
static json_node *field(json_node *n, const adamic_string *key) {
  for (size_t i = 0; i < n->count; i++) {
    if (adamic_string_equal(n->keys[i], key)) {
      return n->children[i];
    }
  }
  return NULL;
}
static const char *kind_name(const json_node *n) {
  if (n == NULL) {
    return "undefined";
  }
  switch (n->kind) {
  case json_number:
    return "number";
  case json_string:
    return "string";
  case json_boolean:
    return "boolean";
  case json_null:
    return "null";
  case json_array:
    return "array";
  default:
    return "object";
  }
}
static bool matches(json_node *n, const adamic_json_parse_schema *s) {
  if (s->kind == json_raw) {
    return true;
  }
  if (n == NULL) {
    return s->optional || s->kind == json_undefined;
  }
  if (s->kind == json_union) {
    for (size_t i = 0; i < s->count; i++) {
      if (matches(n, s->members[i])) {
        return true;
      }
    }
    return false;
  }
  if (n->kind != s->kind) {
    return false;
  }
  if (s->literal != NULL) {
    if (s->kind == json_string) {
      return adamic_string_equal(n->value.reference, s->literal);
    }
    if (s->kind == json_number) {
      return n->value.number == adamic_number_parse_float(s->literal);
    }
    if (s->kind == json_boolean) {
      return n->value.boolean == (s->literal->length == 4);
    }
  }
  return true;
}
static _Noreturn void boundary(json_node *n, const adamic_json_parse_schema *s,
                               adamic_string *path) {
  adamic_string *m = join(text("boundary check: "), adamic_retain(path));
  m = join(m, text(" expected "));
  m = join(m, adamic_retain(s->name));
  m = join(m, text(", got "));
  m = join(m, text(kind_name(n)));
  adamic_panic(m->bytes, m->length);
}
static void validate(json_node *n, const adamic_json_parse_schema *s,
                     adamic_string *path) {
  if (!matches(n, s)) {
    boundary(n, s, path);
  }
  if (n == NULL || s->kind == json_raw) {
    return;
  }
  if (s->kind == json_union) {
    for (size_t i = 0; i < s->count; i++) {
      if (matches(n, s->members[i])) {
        validate(n, s->members[i], path);
        return;
      }
    }
  }
  if (s->kind == json_object) {
    for (size_t i = 0; i < s->count; i++) {
      adamic_string *next = join(join(adamic_retain(path), text(".")),
                                 adamic_retain(s->fields[i].name));
      validate(field(n, s->fields[i].name), s->fields[i].schema, next);
      adamic_release(next);
    }
  }
  if (s->kind == json_array) {
    for (size_t i = 0; i < n->count; i++) {
      char index[40];
      snprintf(index, sizeof index, "[%zu]", i);
      adamic_string *next = join(adamic_retain(path), text(index));
      validate(n->children[i], s->element, next);
      adamic_release(next);
    }
  }
}
static adamic_value materialize(json_node *n,
                                const adamic_json_parse_schema *s) {
  adamic_value r = {0};
  if (s->kind == json_union) {
    for (size_t i = 0; i < s->count; i++) {
      if (matches(n, s->members[i])) {
        r = materialize(n, s->members[i]);
        break;
      }
    }
  } else if (n->kind == json_string) {
    r.reference = adamic_retain(n->value.reference);
  } else if (n->kind == json_number || n->kind == json_boolean) {
    r = n->value;
  } else if (n->kind == json_null) {
    r.reference = s->of == 10 ? &adamic_null : NULL;
  } else if (n->kind == json_array) {
    adamic_array *a =
        adamic_array_new(n->count, s->element->of != 1 && s->element->of != 2 &&
                                       s->element->of != 7);
    for (size_t i = 0; i < n->count; i++) {
      adamic_array_push(a, materialize(n->children[i], s->element));
    }
    r.reference = a;
  } else if (n->kind == json_object) {
    adamic_object *o = adamic_object_new(s->shape);
    for (size_t i = 0; i < s->count; i++) {
      json_node *child = field(n, s->fields[i].name);
      o->slots[i] = materialize(child, s->fields[i].schema);
    }
    r.reference = o;
  }
  if (s->of == 10) {
    if (n->kind == json_null) {
      r.reference = &adamic_null;
    } else if (n->kind == json_number) {
      r.reference = adamic_box_number(n->value.number);
    } else if (n->kind == json_boolean) {
      r.reference = n->value.boolean ? &adamic_box_true : &adamic_box_false;
    }
  }
  if (s->of == 7 && n->kind == json_number) {
    r.number =
        adamic_maybe_number_pack((adamic_maybe_number){true, n->value.number});
  }
  return r;
}
static void dispose(json_parser *p) {
  for (size_t i = 0; i < p->depth; i++) {
    adamic_release(p->stack[i].key);
  }
  free(p->stack);
  free(p->units);
  json_node *n = p->allocated;
  while (n != NULL) {
    json_node *next = n->next;
    if (n->kind == json_string) {
      adamic_release(n->value.reference);
    }
    for (size_t i = 0; i < n->count; i++) {
      if (n->kind == json_object) {
        adamic_release(n->keys[i]);
      }
    }
    free(n->children);
    free(n->keys);
    free(n);
    n = next;
  }
}
adamic_value adamic_json_parse(const adamic_string *input,
                               const adamic_json_parse_schema *check,
                               const adamic_json_parse_schema *layout) {
  json_parser p = {0};
  p.source = input;
  p.length = (size_t)adamic_string_length(input);
  p.units = memory((p.length + 1) * sizeof *p.units);
  for (size_t i = 0; i < p.length; i++) {
    p.units[i] = adamic_string_char_code_at(input, (double)i);
  }
  json_node *root = parse(&p);
  adamic_value result = {0};
  if (!p.failed && check != NULL) {
    adamic_string *path = text("$");
    validate(root, check, path);
    adamic_release(path);
    result = materialize(root, layout);
  }
  dispose(&p);
  return result;
}
