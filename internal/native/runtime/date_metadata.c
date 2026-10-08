// Intrinsic Date descriptors are fixed: prototype mutation remains refused.
#include "adamic.h"
#include <string.h>

static bool date_key_is(const adamic_string *key, const char *name) {
 size_t length = strlen(name);
 return key->length == length && memcmp(key->bytes, name, length) == 0;
}
bool adamic_date_has_own(int target, const adamic_string *key) {
 static const char *const constructor[] = {"length", "name", "prototype", "now", "parse", "UTC"};
 static const char *const prototype[] = {
  "constructor", "toString", "toDateString", "toTimeString", "toLocaleString", "toLocaleDateString", "toLocaleTimeString", "valueOf", "getTime",
  "getFullYear", "getUTCFullYear", "getMonth", "getUTCMonth", "getDate", "getUTCDate", "getDay", "getUTCDay", "getHours", "getUTCHours",
  "getMinutes", "getUTCMinutes", "getSeconds", "getUTCSeconds", "getMilliseconds", "getUTCMilliseconds", "getTimezoneOffset", "getYear",
  "setTime", "setMilliseconds", "setUTCMilliseconds", "setSeconds", "setUTCSeconds", "setMinutes", "setUTCMinutes", "setHours", "setUTCHours",
  "setDate", "setUTCDate", "setMonth", "setUTCMonth", "setFullYear", "setUTCFullYear", "setYear", "toUTCString", "toGMTString", "toISOString", "toJSON"
 };
 if (target == 2) return date_key_is(key, "length") || date_key_is(key, "name");
 const char *const *names = target == 0 ? constructor : prototype;
 size_t count = target == 0 ? sizeof constructor / sizeof constructor[0] : sizeof prototype / sizeof prototype[0];
 for (size_t index = 0; index < count; index++) if (date_key_is(key, names[index])) return true;
 return false;
}
