// Copyright 2011 the V8 project authors. All rights reserved.
// BSD license in THIRD_PARTY_NOTICES.md.
// ParseES5DateTime and Day/Time/TimeZone composers from Node 24.19.0's V8.
// Lowering admits only the ISO grammar with an explicit zone for date-time forms.
// There is no legacy parser or host timezone fallback.
typedef struct { const unsigned char *bytes; size_t at, length; } date_iso_scanner;
static int iso_digit(date_iso_scanner *s, int count) {
 int value = 0;
 for (int i = 0; i < count; i++) {
  if (s->at == s->length || s->bytes[s->at] < '0' || s->bytes[s->at] > '9') return -1;
  value = value * 10 + s->bytes[s->at++] - '0';
 }
 return value;
}
static bool iso_skip(date_iso_scanner *s, unsigned char character) {
 if (s->at == s->length || s->bytes[s->at] != character) return false;
 s->at++; return true;
}
double adamic_date_parse_iso(const adamic_string *text) {
 date_iso_scanner s = {(const unsigned char *)text->bytes,0,text->length};
 int sign = 1, width = 4;
 if (iso_skip(&s,'+')) width = 6;
 else if (iso_skip(&s,'-')) { sign = -1; width = 6; }
 int year = iso_digit(&s,width);
 if (year < 0 || (sign < 0 && year == 0)) return NAN;
 double parts[7] = {year * sign,0,1,0,0,0,0};
 if (iso_skip(&s,'-')) {
  int month = iso_digit(&s,2); if (month < 1 || month > 12) return NAN;
  parts[1] = month - 1;
  if (iso_skip(&s,'-')) {
   int day = iso_digit(&s,2); if (day < 1 || day > 31) return NAN;
   parts[2] = day;
  }
 }
 double offset = 0;
 if (iso_skip(&s,'T') || iso_skip(&s,'t')) {
  int hour = iso_digit(&s,2);
  if (hour < 0 || hour > 24 || !iso_skip(&s,':')) return NAN;
  int minute = iso_digit(&s,2); if (minute < 0 || minute > 59) return NAN;
  parts[3] = hour; parts[4] = minute;
  bool fractional_nonzero = false;
  if (iso_skip(&s,':')) {
   int second = iso_digit(&s,2); if (second < 0 || second > 59) return NAN;
   parts[5] = second;
   if (iso_skip(&s,'.')) {
    // V8 ReadUnsignedNumeral keeps nine significant digits, skipping leading
    // zeros; ReadMilliseconds caps the token width at nine before scaling.
    // Preserve even this leading-zero behavior, rather than rounding the prefix.
    size_t start = s.at, significant = 0; int milliseconds = 0;
    bool leading = true;
    while (s.at < s.length && s.bytes[s.at] >= '0' && s.bytes[s.at] <= '9') {
     int digit = s.bytes[s.at++] - '0';
     fractional_nonzero = fractional_nonzero || digit != 0;
     if (leading && digit == 0) continue;
     leading = false;
     if (significant < 9) milliseconds = milliseconds * 10 + digit;
     significant++;
    }
    if (s.at == start) return NAN;
    size_t digits = s.at - start; if (digits > 9) digits = 9;
    while (digits < 3) { milliseconds *= 10; digits++; }
    while (digits > 3) { milliseconds /= 10; digits--; }
    parts[6] = milliseconds;
   }
  }
  if (hour == 24 && (minute != 0 || parts[5] != 0 || fractional_nonzero)) return NAN;
  if (!(iso_skip(&s,'Z') || iso_skip(&s,'z'))) {
   int zone_sign;
   if (iso_skip(&s,'+')) zone_sign = 1;
   else if (iso_skip(&s,'-')) zone_sign = -1;
   else return NAN;
   int hours = iso_digit(&s,2); if (hours < 0 || hours > 23) return NAN;
   (void)iso_skip(&s,':');
   int minutes = iso_digit(&s,2); if (minutes < 0 || minutes > 59) return NAN;
   offset = (hours * 60 + minutes) * 60000.0 * zone_sign;
  }
 }
 if (s.at != s.length) return NAN;
 return time_clip(make_day(parts[0],parts[1],parts[2]) * 86400000 + make_time(parts+3) - offset);
}
