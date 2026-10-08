// Copyright 2012 the V8 project authors. All rights reserved.
// BSD license in THIRD_PARTY_NOTICES.md.
// UTC portions of Node 24.19.0's V8 date.cc/dateparser-inl.h/builtins-date.cc.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <stdio.h>
#include <string.h>
#include <time.h>

static const adamic_shape date_shape = {0, NULL, NULL, NULL};
static double time_clip(double time) {
 if (!isfinite(time) || fabs(time) > 8640000000000000.0) return NAN;
 return trunc(time) + 0.0;
}
adamic_object *adamic_date_new(double time) {
 adamic_object *date = adamic_allocate(sizeof *date + sizeof date->slots[0], adamic_kind_object);
 // Initialize the current object layout without overwriting the allocator header.
 memset((char *)date + sizeof date->heap, 0, sizeof *date - sizeof date->heap + sizeof date->slots[0]);
 date->shape = &date_shape; date->class = NULL; date->frozen = false;
 date->slots[0].number = time_clip(time);
 return date;
}
double adamic_date_value(const adamic_object *date) { return date->slots[0].number; }

// Floor division is essential before the epoch. All inputs to the integer part are bounded.
static int64_t floor_div(int64_t value, int64_t divisor) {
 int64_t quotient = value / divisor;
 return quotient - (value % divisor < 0);
}
// DateCache::YearMonthDayFromDays, without the cache or local-time adjustment.
static void civil_parts(int64_t day, double *parts) {
 static const int months[] = {31,28,31,30,31,30,31,31,30,31,30,31};
 int64_t days = day + 1005 * 146097 - (30 * 365 + 7);
 int64_t year = 400 * (days / 146097) - 400000;
 days %= 146097;
 days--;
 int64_t centuries = days / 36524; days %= 36524; year += 100 * centuries;
 days++;
 int64_t fours = days / 1461; days %= 1461; year += 4 * fours;
 days--;
 int64_t years = days / 365; days %= 365; year += years;
 bool leap = (!centuries || fours) && !years;
 days += leap;
 int month = 0;
 while (month < 11) {
  int length = months[month] + (month == 1 && leap ? 1 : 0);
  if (days < length) break;
  days -= length; month++;
 }
 parts[0] = (double)year; parts[1] = month; parts[2] = (double)days + 1;
}
// MakeDay: truncation precedes month normalization, and bounds precede integer conversion.
static double make_day(double year, double month, double day) {
 if (!isfinite(year) || !isfinite(month) || !isfinite(day) || year < -1000000 || year > 1000000 || month < -10000000 || month > 10000000) return NAN;
 int64_t y = (int64_t)trunc(year), m = (int64_t)trunc(month);
 y += floor_div(m,12); m -= floor_div(m,12) * 12;
 const int64_t delta = 399999;
 const int64_t base = 365 * (1970 + delta) + (1970 + delta) / 4 - (1970 + delta) / 100 + (1970 + delta) / 400;
 int64_t days = 365 * (y + delta) + (y + delta) / 4 - (y + delta) / 100 + (y + delta) / 400 - base;
 static const int ordinary[] = {0,31,59,90,120,151,181,212,243,273,304,334};
 static const int leap[] = {0,31,60,91,121,152,182,213,244,274,305,335};
 days += (y % 4 != 0 || (y % 100 == 0 && y % 400 != 0)) ? ordinary[m] : leap[m];
 return (double)(days - 1) + trunc(day);
}
static double make_time(const double *parts) {
 for (int i = 0; i < 4; i++) if (!isfinite(parts[i])) return NAN;
 return trunc(parts[0]) * 3600000 + trunc(parts[1]) * 60000 + trunc(parts[2]) * 1000 + trunc(parts[3]);
}
static double make_date(const double *parts) {
 double time = make_time(parts + 3);
 double day = make_day(parts[0], parts[1], parts[2]);
 return time_clip(day * 86400000 + time);
}
double adamic_date_utc(size_t count, const double *arguments) {
 double parts[7] = {NAN, 0, 1, 0, 0, 0, 0};
 for (size_t i = 0; i < count && i < 7; i++) parts[i] = arguments[i];
 double year = trunc(parts[0]);
 if (year >= 0 && year <= 99) parts[0] = year + 1900;
 return make_date(parts);
}
static void split_time(double time, double *parts) {
 int64_t day = (int64_t)floor(time / 86400000);
 civil_parts(day, parts);
 int64_t within = (int64_t)(time - (double)day * 86400000);
 parts[3] = (double)(within / 3600000);
 parts[4] = (double)(within / 60000 % 60);
 parts[5] = (double)(within / 1000 % 60);
 parts[6] = (double)(within % 1000);
}
double adamic_date_get(const adamic_object *date, int field) {
 double time = adamic_date_value(date);
 if (isnan(time)) return NAN;
 double parts[7]; split_time(time, parts);
 if (field == 3) { int64_t day = (int64_t)floor(time / 86400000); return (double)(day - floor_div(day + 4, 7) * 7 + 4); }
 return parts[field < 3 ? field : field - 1];
}
adamic_string *adamic_date_iso(const adamic_object *date) {
 double time = adamic_date_value(date);
 if (isnan(time)) { static const char message[] = "RangeError: Invalid time value"; adamic_panic(message, sizeof message - 1); }
 double parts[7]; split_time(time, parts);
 char text[32]; int year = (int)parts[0];
 int length;
 if (year >= 0 && year <= 9999) {
  length = snprintf(text, sizeof text, "%04d-%02d-%02dT%02d:%02d:%02d.%03dZ", year, (int)parts[1]+1, (int)parts[2], (int)parts[3], (int)parts[4], (int)parts[5], (int)parts[6]);
 } else {
  length = snprintf(text, sizeof text, "%c%06d-%02d-%02dT%02d:%02d:%02d.%03dZ", year < 0 ? '-' : '+', year < 0 ? -year : year, (int)parts[1]+1, (int)parts[2], (int)parts[3], (int)parts[4], (int)parts[5], (int)parts[6]);
 }
 adamic_string *result = adamic_string_allocate((size_t)length);
 memcpy((char *)result->bytes, text, (size_t)length);
 return result;
}
// V8 bootstrapper.cc's Date installation. Mutation/expando writes stay refused.
bool adamic_date_has_own(const adamic_string *key, bool prototype) {
 static const char *const constructor_names[] = {"length", "name", "prototype", "now", "parse", "UTC"};
 static const char *const prototype_names[] = {"constructor", "toString", "toDateString", "toTimeString", "toISOString", "toUTCString", "toGMTString", "getDate", "setDate", "getDay", "getFullYear", "setFullYear", "getHours", "setHours", "getMilliseconds", "setMilliseconds", "getMinutes", "setMinutes", "getMonth", "setMonth", "getSeconds", "setSeconds", "getTime", "setTime", "getTimezoneOffset", "getUTCDate", "setUTCDate", "getUTCDay", "getUTCFullYear", "setUTCFullYear", "getUTCHours", "setUTCHours", "getUTCMilliseconds", "setUTCMilliseconds", "getUTCMinutes", "setUTCMinutes", "getUTCMonth", "setUTCMonth", "getUTCSeconds", "setUTCSeconds", "valueOf", "getYear", "setYear", "toJSON", "toLocaleString", "toLocaleDateString", "toLocaleTimeString"};
 const char *const *names = prototype ? prototype_names : constructor_names;
 size_t count = prototype ? sizeof prototype_names / sizeof *prototype_names : sizeof constructor_names / sizeof *constructor_names;
 for (size_t i = 0; i < count; i++) {
  if (key->length == strlen(names[i]) && memcmp(key->bytes, names[i], key->length) == 0) return true;
 }
 return false;
}

// Node's CurrentTimeValue floors its system clock to whole epoch milliseconds.
// clock_gettime(CLOCK_REALTIME) is provided by macOS, Linux and wasi-libc.
double adamic_date_now(void) {
 struct timespec time;
 if (clock_gettime(CLOCK_REALTIME, &time) != 0) {
  static const char message[] = "Date.now: CLOCK_REALTIME is unavailable";
  adamic_panic(message, sizeof message - 1);
 }
 return (double)time.tv_sec * 1000.0 + (double)(time.tv_nsec / 1000000);
}
static adamic_value date_now_method(adamic_closure *self, adamic_value *arguments) {
 (void)self; (void)arguments;
 return (adamic_value){.number = adamic_date_now()};
}
static adamic_closure date_now_closure = {{0, adamic_kind_closure, 0}, date_now_method, 0};
adamic_closure *adamic_date_now_function(void) { return &date_now_closure; }

adamic_string *adamic_date_utc_string(const adamic_object *date) {
 static const char *const days[] = {"Sun","Mon","Tue","Wed","Thu","Fri","Sat"};
 static const char *const months[] = {"Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"};
 double time = adamic_date_value(date);
 if (isnan(time)) { static adamic_string invalid = ADAMIC_STRING("Invalid Date"); return &invalid; }
 double parts[7]; split_time(time,parts);
 char text[64]; int year = (int)parts[0];
 int length = snprintf(text,sizeof text,year < 0 ? "%s, %02d %s %05d %02d:%02d:%02d GMT" : "%s, %02d %s %04d %02d:%02d:%02d GMT",
  days[(int)adamic_date_get(date,3)],(int)parts[2],months[(int)parts[1]],year,(int)parts[3],(int)parts[4],(int)parts[5]);
 adamic_string *result = adamic_string_allocate((size_t)length);
 memcpy((char *)result->bytes,text,(size_t)length);
 return result;
}

#include "date_iso_parse_impl.h"
