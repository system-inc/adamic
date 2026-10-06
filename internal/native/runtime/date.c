// date.c: proleptic Gregorian arithmetic in milliseconds, independent of libc time_t.
#include "adamic.h"
#include <stdio.h>
#include <string.h>

static const adamic_shape date_shape = {0, NULL, NULL, NULL};
static double time_clip(double time) {
 if (!isfinite(time) || fabs(time) > 8640000000000000.0) return NAN;
 return trunc(time) + 0.0;
}
adamic_object *adamic_date_new(double time) {
 adamic_object *date = adamic_allocate(sizeof *date + sizeof date->slots[0], adamic_kind_object);
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
static int64_t civil_day(int64_t year, int month, int day) {
 year -= month <= 2;
 int64_t era = floor_div(year, 400);
 int64_t within = year - era * 400;
 int shifted = month + (month > 2 ? -3 : 9);
 int64_t ordinal = (153 * shifted + 2) / 5 + day - 1;
 return era * 146097 + within * 365 + within / 4 - within / 100 + ordinal - 719468;
}
static void civil_parts(int64_t day, double *parts) {
 int64_t shifted = day + 719468;
 int64_t era = floor_div(shifted, 146097);
 int64_t ordinal = shifted - era * 146097;
 int64_t within = (ordinal - ordinal / 1460 + ordinal / 36524 - ordinal / 146096) / 365;
 int64_t year = within + era * 400;
 int64_t year_day = ordinal - (365 * within + within / 4 - within / 100);
 int64_t month = (5 * year_day + 2) / 153;
 parts[2] = (double)(year_day - (153 * month + 2) / 5 + 1);
 month += month < 10 ? 3 : -9;
 year += month <= 2;
 parts[0] = (double)year; parts[1] = (double)(month - 1);
}
static double make_day(double year, double month, double day) {
 // V8 uses these conservative bounds before normalizing the month.
 if (!isfinite(year) || !isfinite(month) || !isfinite(day) || year < -1000000 || year > 1000000 || month < -10000000 || month > 10000000) return NAN;
 int64_t y = (int64_t)trunc(year), m = (int64_t)trunc(month);
 y += floor_div(m, 12); m -= floor_div(m, 12) * 12;
 return (double)civil_day(y, (int)m + 1, 1) + trunc(day) - 1;
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
 if (field == 8) return 0;
 if (field == 9) return parts[0] - 1900;
 return parts[field < 3 ? field : field - 1];
}
double adamic_date_set(adamic_object *date, int field, size_t count, const double *arguments) {
 double time = adamic_date_value(date);
 double first = count == 0 ? NAN : arguments[0];
 if (field == 10) { date->slots[0].number = time_clip(first); return date->slots[0].number; }
 if (isnan(time) && field != 0 && field != 9) return NAN;
 if (isnan(time)) time = 0;
 double parts[7]; split_time(time, parts);
 if (field == 9) {
  double year = trunc(first); parts[0] = year >= 0 && year <= 99 ? year + 1900 : first;
 } else {
  int start = field < 3 ? field : field - 1;
  int end = field == 0 || field == 1 ? 3 : field == 2 ? 3 : 7;
  parts[start] = first;
  for (size_t i = 1; i < count && start + (int)i < end; i++) parts[start + (int)i] = arguments[i];
 }
 date->slots[0].number = make_date(parts);
 return date->slots[0].number;
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
static int digits(const adamic_string *text, size_t *position, int width) {
 int value = 0;
 for (int i = 0; i < width; i++) {
  if (*position >= text->length) return -1;
  char c = text->bytes[(*position)++]; if (c < '0' || c > '9') return -1;
  value = value * 10 + c - '0';
 }
 return value;
}
static bool take(const adamic_string *text, size_t *position, char c) {
 if (*position == text->length || text->bytes[*position] != c) return false;
 (*position)++; return true;
}
// These are the two legacy forms Date itself prints with TZ=UTC, also checked by test262.
static double parse_rendered(const adamic_string *text) {
 if (text->length >= 128) return NAN;
 char input[128]; memcpy(input, text->bytes, text->length); input[text->length] = '\0';
 char weekday[4], month[4]; int year, day, hour, minute, second, consumed = 0;
 int fields = sscanf(input, "%3s, %d %3s %d %d:%d:%d GMT%n", weekday, &day, month, &year, &hour, &minute, &second, &consumed);
 if (fields != 7 || consumed == 0 || (size_t)consumed != text->length) {
  consumed = 0;
  fields = sscanf(input, "%3s %3s %d %d %d:%d:%d GMT+0000 (Coordinated Universal Time)%n", weekday, month, &day, &year, &hour, &minute, &second, &consumed);
  if (fields != 7 || consumed == 0 || (size_t)consumed != text->length) return NAN;
 }
 if (year < 0 || day < 1 || day > 31 || hour < 0 || hour > 24 || minute < 0 || minute > 59 || second < 0 || second > 59 || (hour == 24 && (minute != 0 || second != 0))) return NAN;
 static const char *const months[] = {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"};
 int index = 0; while (index < 12 && strcmp(month, months[index]) != 0) index++;
 if (index == 12) return NAN;
 // V8's legacy parser applies its two-digit-year rule even when the zeros were padded.
 if (year < 50) year += 2000; else if (year < 100) year += 1900;
 double parts[7] = {year, index, day, hour, minute, second, 0};
 return make_date(parts);
}
double adamic_date_parse_iso(const adamic_string *text) {
 if (text->length > 0 && text->bytes[0] >= 'A' && text->bytes[0] <= 'Z') return parse_rendered(text);
 size_t at = 0; int sign = 1, width = 4; bool expanded = false;
 if (take(text, &at, '-')) { sign = -1; width = 6; expanded = true; }
 else if (take(text, &at, '+')) { width = 6; expanded = true; }
 int year = digits(text, &at, width);
 if (year < 0 || (expanded && sign < 0 && year == 0)) return NAN;
 double parts[7] = {(double)(sign * year), 0, 1, 0, 0, 0, 0};
 if (take(text, &at, '-')) {
  int month = digits(text, &at, 2); if (month < 1 || month > 12) return NAN; parts[1] = month - 1;
  if (take(text, &at, '-')) { int day = digits(text, &at, 2); if (day < 1 || day > 31) return NAN; parts[2] = day; }
 }
 int offset = 0;
 if (take(text, &at, 'T')) {
  int hour = digits(text, &at, 2); if (hour < 0 || hour > 24 || !take(text, &at, ':')) return NAN;
  int minute = digits(text, &at, 2); if (minute < 0 || minute > 59) return NAN;
  parts[3] = hour; parts[4] = minute;
  if (take(text, &at, ':')) {
   int second = digits(text, &at, 2); if (second < 0 || second > 59) return NAN; parts[5] = second;
   if (take(text, &at, '.')) {
    int millisecond = 0, count = 0; bool nonzero_fraction = false;
    while (at < text->length && text->bytes[at] >= '0' && text->bytes[at] <= '9') {
     nonzero_fraction = nonzero_fraction || text->bytes[at] != '0';
     if (count < 3) millisecond = millisecond * 10 + text->bytes[at] - '0';
     count++; at++;
    }
    if (count == 0 || (hour == 24 && nonzero_fraction)) return NAN;
    while (count < 3) {millisecond *= 10; count++;}
    parts[6] = millisecond;
   }
  }
  if (hour == 24 && (parts[4] != 0 || parts[5] != 0 || parts[6] != 0)) return NAN;
  if (!take(text, &at, 'Z') && at < text->length) {
   int zone_sign;
   if (take(text, &at, '+')) zone_sign = 1;
   else if (take(text, &at, '-')) zone_sign = -1;
   else return NAN;
   int zone_hour = digits(text, &at, 2); if (zone_hour < 0 || zone_hour > 23 || !take(text, &at, ':')) return NAN;
   int zone_minute = digits(text, &at, 2); if (zone_minute < 0 || zone_minute > 59) return NAN;
   offset = zone_sign * (zone_hour * 60 + zone_minute) * 60000;
  }
 }
 if (at != text->length) return NAN;
 // Clip after applying the timezone, so an out-of-range local time may still be valid UTC.
 return time_clip(make_day(parts[0], parts[1], parts[2]) * 86400000 + make_time(parts + 3) - offset);
}

// Stable UTC renderings use V8's English month/day names, with the oracle's fixed TZ=UTC.
adamic_string *adamic_date_format(const adamic_object *date, int style) {
 static const char *const days[] = {"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"};
 static const char *const months[] = {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"};
 char text[128];
 double time = adamic_date_value(date);
 if (isnan(time)) {
  memcpy(text, "Invalid Date", 13);
 } else {
  double parts[7]; split_time(time, parts);
  int year = (int)parts[0];
  char year_text[16];
  if (year < 0) snprintf(year_text, sizeof year_text, "-%04d", -year);
  else snprintf(year_text, sizeof year_text, "%04d", year);
  const char *weekday = days[(int)adamic_date_get(date, 3)];
  const char *month = months[(int)parts[1]];
  char day_text[48], time_text[72];
  snprintf(day_text, sizeof day_text, "%s %s %02d %s", weekday, month, (int)parts[2], year_text);
  snprintf(time_text, sizeof time_text, "%02d:%02d:%02d GMT+0000 (Coordinated Universal Time)", (int)parts[3], (int)parts[4], (int)parts[5]);
  if (style == 1) snprintf(text, sizeof text, "%s, %02d %s %s %02d:%02d:%02d GMT", weekday, (int)parts[2], month, year_text, (int)parts[3], (int)parts[4], (int)parts[5]);
  else if (style == 2) snprintf(text, sizeof text, "%s", day_text);
  else if (style == 3) snprintf(text, sizeof text, "%s", time_text);
  else snprintf(text, sizeof text, "%s %s", day_text, time_text);
 }
 size_t length = strlen(text);
 adamic_string *result = adamic_string_allocate(length);
 memcpy((char *)result->bytes, text, length);
 return result;
}
