// Math integer and float conversions follow V8 13.6.233.17, src/builtins/math.tq and
// src/numbers/conversions-inl.h (DoubleToInt32, DoubleToFloat32).
// Copyright the V8 project authors. BSD-3-Clause; see THIRD_PARTY_NOTICES.md.
#include "adamic.h"
#include <float.h>
#include <math.h>
#include <stdint.h>
#include <string.h>

// V8 extracts the significand and shifts it rather than calling libm. Unsigned arithmetic
// expresses the low 32 bits without signed overflow or out-of-range floating conversions.
static uint32_t library_uint32(double value) {
 uint64_t bits;
 memcpy(&bits, &value, sizeof bits);
 unsigned encoded = (unsigned)((bits >> 52) & 0x7ff);
 if (encoded == 0 || encoded == 0x7ff) return 0;
 int exponent = (int)encoded - 1023 - 52;
 uint64_t significand = (bits & UINT64_C(0xfffffffffffff)) | UINT64_C(0x10000000000000);
 uint32_t result;
 if (exponent < 0) {
  if (exponent <= -53) return 0;
  result = (uint32_t)(significand >> -exponent);
 } else {
  if (exponent > 31) return 0;
  result = (uint32_t)(significand << exponent);
 }
 return (bits >> 63) ? (uint32_t)(0u - result) : result;
}

double adamic_math_clz32(double value) {
 uint32_t bits = library_uint32(value);
 unsigned count = 0;
 if (bits == 0) return 32;
 while ((bits & UINT32_C(0x80000000)) == 0) { count++; bits <<= 1; }
 return (double)count;
}

double adamic_math_imul(double left, double right) {
 uint32_t bits = library_uint32(left) * library_uint32(right);
 return bits >= UINT32_C(0x80000000) ? (double)bits - 4294967296.0 : (double)bits;
}

double adamic_math_fround(double value) {
 // V8 guards the conversion at the largest double rounding down to FLT_MAX.
 if (value > FLT_MAX) return value <= 3.4028235677973362e38 ? FLT_MAX : INFINITY;
 if (value < -FLT_MAX) return value >= -3.4028235677973362e38 ? -FLT_MAX : -INFINITY;
 return (double)(float)value;
}

static int library_digit(unsigned char c) {
 if (c >= '0' && c <= '9') return c - '0';
 if (c >= 'a' && c <= 'f') return c - 'a' + 10;
 if (c >= 'A' && c <= 'F') return c - 'A' + 10;
 return 99;
}

double adamic_number_from_string(const adamic_string *text) {
 if (text == NULL) return NAN;
 adamic_string *trimmed = adamic_string_trim_sides((adamic_string *)text, true, true);
 const unsigned char *start = (const unsigned char *)trimmed->bytes;
 const unsigned char *end = start + trimmed->length, *cursor = start;
 double result = NAN;
 if (cursor == end) { result = 0; goto done; }
 if (end - start >= 2 && start[0] == '0') {
  int radix = start[1] == 'x' || start[1] == 'X' ? 16 : start[1] == 'o' || start[1] == 'O' ? 8 : start[1] == 'b' || start[1] == 'B' ? 2 : 0;
  if (radix != 0) {
   cursor += 2;
   if (cursor == end) goto done;
   for (const unsigned char *digit = cursor; digit != end; digit++) if (library_digit(*digit) >= radix) goto done;
   // The existing V8 power-of-two parser supplies exact rounding, including huge inputs.
   adamic_string *digits = adamic_string_slice(trimmed, 2, (double)trimmed->length, true);
   result = adamic_number_parse_int(digits, radix);
   adamic_release(digits);
   goto done;
  }
 }
 if (*cursor == '+' || *cursor == '-') cursor++;
 if (end - cursor == 8 && memcmp(cursor, "Infinity", 8) == 0) {
  result = start[0] == '-' ? -INFINITY : INFINITY; goto done;
 }
 size_t digits = 0;
 while (cursor != end && *cursor >= '0' && *cursor <= '9') { cursor++; digits++; }
 if (cursor != end && *cursor == '.') {
  cursor++;
  while (cursor != end && *cursor >= '0' && *cursor <= '9') { cursor++; digits++; }
 }
 if (digits == 0) goto done;
 if (cursor != end && (*cursor == 'e' || *cursor == 'E')) {
  cursor++;
  if (cursor != end && (*cursor == '+' || *cursor == '-')) cursor++;
  const unsigned char *exponent = cursor;
  while (cursor != end && *cursor >= '0' && *cursor <= '9') cursor++;
  if (cursor == exponent) goto done;
 }
 if (cursor == end) result = adamic_number_parse_float(trimmed);
done:
 adamic_release(trimmed);
 return result;
}

double adamic_number_from_union(const adamic_heap *value) {
 if (value == NULL) return NAN;
 switch (value->kind) {
 case adamic_kind_number: return ((const adamic_number_box *)value)->number;
 case adamic_kind_boolean: return ((const adamic_boolean_box *)value)->boolean ? 1 : 0;
 case adamic_kind_string: return adamic_number_from_string((const adamic_string *)value);
 default: {
  static const char message[] = "Number conversion received an unproved object";
  adamic_panic(message, sizeof message - 1);
 }
 }
}

bool adamic_number_has_own_property(const adamic_string *key, bool prototype) {
 static const char *const names[] = {"length", "name", "prototype", "MAX_VALUE", "MIN_VALUE", "NaN", "NEGATIVE_INFINITY", "POSITIVE_INFINITY", "MAX_SAFE_INTEGER", "MIN_SAFE_INTEGER", "EPSILON", "isFinite", "isInteger", "isNaN", "isSafeInteger", "parseFloat", "parseInt"};
 static const char *const prototype_names[] = {"constructor", "toExponential", "toFixed", "toPrecision", "toString", "toLocaleString", "valueOf"};
 const char *const *selected = prototype ? prototype_names : names;
 size_t count = prototype ? sizeof prototype_names / sizeof prototype_names[0] : sizeof names / sizeof names[0];
 if (key == NULL) return false;
 for (size_t index = 0; index < count; index++) {
  if (key->length == strlen(selected[index]) && memcmp(key->bytes, selected[index], key->length) == 0) return true;
 }
 return false;
}
