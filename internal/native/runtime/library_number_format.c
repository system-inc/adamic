// Copyright 2019 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in the
// LICENSE file. See THIRD_PARTY_NOTICES.md.
// Number formatting validation order, from V8 src/builtins/builtins-number.cc and number.tq.
// Successful calls reuse the existing V8 digit and radix ports without changing their bits.
#include "adamic.h"
#include <math.h>

static adamic_string *format_range_error(const char *message) {
 adamic_string text = {{0, adamic_kind_string, 0}, strlen(message), message, 0, NULL, NULL, 0};
 adamic_string *owned = adamic_string_concat(1, (adamic_string *const[]){&text});
 adamic_object *error = adamic_error_new(owned);
 adamic_release(owned);
 static adamic_string name = ADAMIC_STRING("RangeError");
 adamic_release(error->slots[0].reference);
 error->slots[0].reference = adamic_retain(&name);
 adamic_error_tag(error);
 adamic_thrown = error;
 return NULL;
}

static double format_integer(double value) {
 return isnan(value) ? 0 : trunc(value);
}

adamic_string *adamic_number_checked_fixed(double value, double digits) {
 double integer = format_integer(digits);
 if (integer < 0 || integer > 100) {
  return format_range_error("toFixed() digits argument must be between 0 and 100");
 }
 return adamic_number_to_fixed(value, integer);
}

adamic_string *adamic_number_checked_exponential(double value, double digits) {
 double integer = format_integer(digits);
 // Nonfinite receivers win over an out-of-range fraction count.
 if (isfinite(value) && (integer < 0 || integer > 100)) {
  return format_range_error("toExponential() argument must be between 0 and 100");
 }
 return adamic_number_to_exponential(value, integer, true);
}

adamic_string *adamic_number_checked_precision(double value, double digits) {
 double integer = format_integer(digits);
 if (isfinite(value) && (integer < 1 || integer > 100)) {
  return format_range_error("toPrecision() argument must be between 1 and 100");
 }
 return adamic_number_to_precision(value, integer, true);
}

adamic_string *adamic_number_checked_radix(double value, double radix) {
 double integer = format_integer(radix);
 if (integer < 2 || integer > 36) {
  return format_range_error("toString() radix argument must be between 2 and 36");
 }
 return adamic_number_to_radix(value, integer);
}
