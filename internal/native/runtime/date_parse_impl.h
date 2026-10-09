// Date parser adapted from V8's dateparser.{h,cc} and dateparser-inl.h.
// Copyright 2011 the V8 project authors. All rights reserved.
// BSD license: date_v8_license.txt. UTC is the compiler/oracle's fixed local zone.

enum date_token_kind {date_end, date_number, date_symbol, date_word, date_space, date_unknown, date_invalid};
enum date_keyword {date_garbage, date_month, date_zone, date_separator, date_ampm};
typedef struct {int kind, value, keyword; size_t length;} date_token;
typedef struct {const adamic_string *text; size_t at, length; date_token next;} date_scanner;
typedef struct {
 int days[3], times[4], day_count, time_count, month, ampm, zone_sign, zone_hour, zone_minute;
 bool iso;
} date_composer;
static unsigned date_character(const date_scanner *s) {
 return s->at >= s->length ? 0 : (unsigned)adamic_string_char_code_at(s->text, (double)s->at);
}
static bool date_whitespace(unsigned c) {
 return c == 0x20 || c == 0x09 || c == 0x0b || c == 0x0c || c == 0xa0 || c == 0xfeff || c == 0x1680 ||
 (c >= 0x2000 && c <= 0x200a) || c == 0x202f || c == 0x205f || c == 0x3000;
}
static date_token date_scan(date_scanner *s) {
 unsigned c = date_character(s); size_t start = s->at;
 if (c == 0) return (date_token){date_end, 0, 0, 0};
 if (c >= '0' && c <= '9') {
  int value = 0, significant = 0;
  while (date_character(s) == '0') s->at++;
  while ((c = date_character(s)) >= '0' && c <= '9') {
   if (significant < 9) value = value * 10 + (int)c - '0';
   significant++; s->at++;
  }
  return (date_token){date_number, value, 0, s->at - start};
 }
 if (c == ':' || c == '-' || c == '+' || c == '.' || c == ')') {
  s->at++; return (date_token){date_symbol, (int)c, 0, 1};
 }
 if (c >= 'A' && !date_whitespace(c)) {
  unsigned prefix[3] = {0, 0, 0}; size_t length = 0;
  while ((c = date_character(s)) >= 'A' && !date_whitespace(c)) {
   if (length < 3) prefix[length] = c >= 'A' && c <= 'Z' ? c + 32 : c;
   length++; s->at++;
  }
  static const char *const words[] = {"jan","feb","mar","apr","may","jun","jul","aug","sep","oct","nov","dec","am","pm","ut","utc","z","gmt","cdt","cst","edt","est","mdt","mst","pdt","pst","t"};
  static const int values[] = {1,2,3,4,5,6,7,8,9,10,11,12,0,12,0,0,0,0,-5,-6,-4,-5,-6,-7,-7,-8,0};
  for (size_t i = 0; i < sizeof words / sizeof words[0]; i++) {
   bool match = true;
   for (size_t j = 0; j < 3; j++) {
    unsigned expected = j < strlen(words[i]) ? (unsigned char)words[i][j] : 0;
    if (prefix[j] != expected) match = false;
   }
   if (match && (length <= 3 || i < 12)) {
    int keyword = i < 12 ? date_month : i < 14 ? date_ampm : i == 26 ? date_separator : date_zone;
    return (date_token){date_word, values[i], keyword, length};
   }
  }
  return (date_token){date_word, 0, date_garbage, length};
 }
 if (date_whitespace(c) || c == 0x0a || c == 0x0d || c == 0x2028 || c == 0x2029) {
  s->at++; return (date_token){date_space, 0, 0, 1};
 }
 if (c == '(') {
  int balance = 0;
  do {
   c = date_character(s); if (c == '(') balance++; else if (c == ')') balance--;
   s->at++;
  } while (balance > 0 && date_character(s) != 0);
 } else s->at++;
 return (date_token){date_unknown, 0, 0, s->at - start};
}
static date_token date_next(date_scanner *s) {date_token t = s->next; s->next = date_scan(s); return t;}
static bool date_skip(date_scanner *s, int c) {
 if (s->next.kind != date_symbol || s->next.value != c) return false;
 date_next(s); return true;
}
static bool date_fixed(date_token t, int width) {return t.kind == date_number && t.length == (size_t)width;}
static bool date_between(int n, int lo, int hi) {return n >= lo && n <= hi;}
static bool date_sign(date_token t) {return t.kind == date_symbol && (t.value == '+' || t.value == '-');}
static bool date_z(date_token t) {return t.kind == date_word && t.keyword == date_zone && t.length == 1 && t.value == 0;}
static bool date_day_add(date_composer *d, int n) {
 if (d->day_count == 3) return false;
 d->days[d->day_count++] = n; return true;
}
static bool date_time_add(date_composer *d, int n) {
 if (d->time_count == 4) return false;
 d->times[d->time_count++] = n; return true;
}
static bool date_time_expecting(const date_composer *d, int n) {
 return ((d->time_count == 1 || d->time_count == 2) && date_between(n,0,59)) || (d->time_count == 3 && date_between(n,0,999));
}
static void date_time_final(date_composer *d, int n) {
 date_time_add(d,n); while (d->time_count < 4) date_time_add(d,0);
}
static int date_milliseconds(date_token t) {
 int length = t.length > 9 ? 9 : (int)t.length;
 int n = t.value;
 while (length < 3) {n *= 10; length++;}
 while (length > 3) {n /= 10; length--;}
 return n;
}
static void date_zone_set(date_composer *d, int hours) {
 d->zone_sign = hours < 0 ? -1 : 1; d->zone_hour = hours * d->zone_sign; d->zone_minute = 0;
}
static date_token date_es5(date_scanner *s, date_composer *d) {
 date_token invalid = {date_invalid,0,0,0};
 if (date_sign(s->next)) {
  date_token sign = date_next(s);
  if (!date_fixed(s->next,6)) return sign;
  int year = date_next(s).value;
  if (sign.value == '-' && year == 0) return sign;
  date_day_add(d, sign.value == '-' ? -year : year);
 } else if (date_fixed(s->next,4)) date_day_add(d,date_next(s).value);
 else return date_next(s);
 if (date_skip(s,'-')) {
  if (!date_fixed(s->next,2) || !date_between(s->next.value,1,12)) return date_next(s);
  date_day_add(d,date_next(s).value);
  if (date_skip(s,'-')) {
   if (!date_fixed(s->next,2) || !date_between(s->next.value,1,31)) return date_next(s);
   date_day_add(d,date_next(s).value);
  }
 }
 if (s->next.kind != date_word || s->next.keyword != date_separator) {
  if (s->next.kind != date_end) return date_next(s);
 } else {
  date_next(s);
  if (!date_fixed(s->next,2) || !date_between(s->next.value,0,24)) return invalid;
  bool midnight = s->next.value == 24;
  date_time_add(d,date_next(s).value);
  if (!date_skip(s,':')) return invalid;
  if (!date_fixed(s->next,2) || !date_between(s->next.value,0,59) || (midnight && s->next.value > 0)) return invalid;
  date_time_add(d,date_next(s).value);
  if (date_skip(s,':')) {
   if (!date_fixed(s->next,2) || !date_between(s->next.value,0,59) || (midnight && s->next.value > 0)) return invalid;
   date_time_add(d,date_next(s).value);
   if (date_skip(s,'.')) {
    if (s->next.kind != date_number || (midnight && s->next.value > 0)) return invalid;
    date_time_add(d,date_milliseconds(date_next(s)));
   }
  }
  if (date_z(s->next)) {date_next(s); date_zone_set(d,0);}
  else if (date_sign(s->next)) {
   d->zone_sign = date_next(s).value == '+' ? 1 : -1;
   if (date_fixed(s->next,4)) {
    int hm = date_next(s).value;
    if (!date_between(hm/100,0,23) || !date_between(hm%100,0,59)) return invalid;
    d->zone_hour = hm/100; d->zone_minute = hm%100;
   } else {
    if (!date_fixed(s->next,2) || !date_between(s->next.value,0,23)) return invalid;
    d->zone_hour = date_next(s).value;
    if (!date_skip(s,':') || !date_fixed(s->next,2) || !date_between(s->next.value,0,59)) return invalid;
    d->zone_minute = date_next(s).value;
   }
  }
  if (s->next.kind != date_end) return invalid;
 }
 if (d->zone_hour < 0 && d->time_count == 0) date_zone_set(d,0);
 d->iso = true; return (date_token){date_end,0,0,0};
}
double adamic_date_parse_iso(const adamic_string *text) {
 date_scanner s = {.text=text, .at=0, .length=(size_t)adamic_string_length(text)};
 s.next = date_scan(&s);
 date_composer d = {.days={0}, .times={0}, .day_count=0, .time_count=0, .month=-1, .ampm=-1, .zone_sign=0, .zone_hour=-1, .zone_minute=-1, .iso=false};
 date_token t = date_es5(&s,&d);
 if (t.kind == date_invalid) return NAN;
 bool read_number = d.day_count != 0;
 for (; t.kind != date_end; t = date_next(&s)) {
  if (t.kind == date_number) {
   read_number = true; int n = t.value;
   if (date_skip(&s,':')) {
    if (date_skip(&s,':')) {
     if (d.time_count != 0) return NAN;
     date_time_add(&d,n); date_time_add(&d,0);
    } else {
     if (!date_time_add(&d,n)) return NAN;
     date_skip(&s,'.');
    }
   } else if (date_skip(&s,'.') && date_time_expecting(&d,n)) {
    date_time_add(&d,n);
    if (s.next.kind != date_number) return NAN;
    date_time_final(&d,date_milliseconds(date_next(&s)));
   } else if (d.zone_hour >= 0 && d.zone_minute < 0 && date_between(n,0,59)) d.zone_minute = n;
   else if (date_time_expecting(&d,n)) {
    date_time_final(&d,n);
    if (s.next.kind != date_end && s.next.kind != date_space && !date_z(s.next) && !date_sign(s.next)) return NAN;
   } else {
    if (!date_day_add(&d,n)) return NAN;
    date_skip(&s,'-');
   }
  } else if (t.kind == date_word) {
   if (t.keyword == date_ampm && d.time_count != 0) d.ampm = t.value;
   else if (t.keyword == date_month) {d.month=t.value; date_skip(&s,'-');}
   else if (t.keyword == date_zone && read_number) date_zone_set(&d,t.value);
   else if (read_number || s.next.kind == date_number) return NAN;
  } else if (date_sign(t) && ((d.zone_hour == 0 && d.zone_minute == 0) || d.time_count != 0)) {
   d.zone_sign = t.value == '+' ? 1 : -1;
   int n=0; size_t length=0;
   if (s.next.kind == date_number) {date_token zone=date_next(&s); n=zone.value; length=zone.length;}
   read_number=true;
   if (s.next.kind == date_symbol && s.next.value == ':') {d.zone_hour=n; d.zone_minute=-1;}
   else if (length == 1 || length == 2) {d.zone_hour=n; d.zone_minute=0;}
   else if (length == 3 || length == 4) {d.zone_hour=n/100; d.zone_minute=n%100;}
   else return NAN;
  } else if ((date_sign(t) || (t.kind == date_symbol && t.value == ')')) && read_number) return NAN;
 }
 if (d.day_count == 0) return NAN;
 while (d.day_count < 3) d.days[d.day_count++]=1;
 int year=0, month, day;
 if (d.month < 0) {
  if (d.iso || !date_between(d.days[0],1,31)) {year=d.days[0]; month=d.days[1]; day=d.days[2];}
  else {month=d.days[0]; day=d.days[1]; year=d.days[2];}
 } else {
  month=d.month;
  if (!date_between(d.days[0],1,31)) {year=d.days[0]; day=d.days[1];}
  else {day=d.days[0]; year=d.days[1];}
 }
 if (!d.iso) {if (date_between(year,0,49)) year+=2000; else if (date_between(year,50,99)) year+=1900;}
 if (!date_between(month,1,12) || !date_between(day,1,31)) return NAN;
 while (d.time_count < 4) d.times[d.time_count++]=0;
 if (d.ampm >= 0) {
  if (!date_between(d.times[0],0,12)) return NAN;
  d.times[0]=d.times[0]%12+d.ampm;
 }
 if (!date_between(d.times[0],0,23) || !date_between(d.times[1],0,59) || !date_between(d.times[2],0,59) || !date_between(d.times[3],0,999)) {
  if (d.times[0]!=24 || d.times[1]!=0 || d.times[2]!=0 || d.times[3]!=0) return NAN;
 }
 double offset=0;
 if (d.zone_sign != 0) {
  unsigned hours=d.zone_hour < 0 ? 0U : (unsigned)d.zone_hour;
  unsigned minutes=d.zone_minute < 0 ? 0U : (unsigned)d.zone_minute;
  unsigned total=hours*3600U+minutes*60U;
  if (total > 2147483647U) return NAN;
  offset=(double)total*d.zone_sign*1000;
 }
 double parts[7]={year,month-1,day,d.times[0],d.times[1],d.times[2],d.times[3]};
 return time_clip(make_day(parts[0],parts[1],parts[2])*86400000+make_time(parts+3)-offset);
}
