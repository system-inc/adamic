// Calendar splitting and English labels reused from library-date/date.c.
// ICU English long names are admitted only for the explicitly supported zone IDs.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#ifndef ADAMIC_TARGET_WASI
#include <time.h>
#include <unistd.h>
#endif

static adamic_string *date_string_text(const char *text) {
 size_t length = strlen(text);
 adamic_string *result = adamic_string_allocate(length);
 memcpy((char *)result->bytes, text, length);
 return result;
}
static adamic_string *date_string_refuse(const char *reason) {
 adamic_string *message = date_string_text(reason);
 adamic_object *error = adamic_error_new(message);
 adamic_release(message);
 static adamic_string name = ADAMIC_STRING("DateStringNotYet");
 adamic_release(error->slots[0].reference);
 error->slots[0].reference = adamic_retain(&name);
 adamic_thrown = error;
 return NULL;
}
static int64_t date_string_floor(int64_t value, int64_t divisor) {
 return value / divisor - (value % divisor < 0);
}
static void date_string_parts(int64_t seconds, int *parts) {
 int64_t day = date_string_floor(seconds,86400);
 int64_t shifted=day+719468, era=date_string_floor(shifted,146097);
 int64_t ordinal=shifted-era*146097;
 int64_t within=(ordinal-ordinal/1460+ordinal/36524-ordinal/146096)/365;
 int64_t year=within+era*400;
 int64_t year_day=ordinal-(365*within+within/4-within/100);
 int64_t month=(5*year_day+2)/153;
 parts[2]=(int)(year_day-(153*month+2)/5+1);
 month+=month<10?3:-9; year+=month<=2;
 parts[0]=(int)year; parts[1]=(int)month-1;
 int64_t rest=seconds-day*86400;
 parts[3]=(int)(rest/3600); parts[4]=(int)(rest/60%60); parts[5]=(int)(rest%60);
 parts[6]=(int)(day-date_string_floor(day+4,7)*7+4);
}

adamic_string *adamic_fs_file_date_string(const adamic_object *date) {
 double time=adamic_fs_file_date_time(date);
 if (isnan(time)) return date_string_text("Invalid Date");
 const char *zone=getenv("TZ");
#ifndef ADAMIC_TARGET_WASI
 char system_zone[512];
 if (zone==NULL) {
  ssize_t length=readlink("/etc/localtime",system_zone,sizeof system_zone-1);
  if (length>0) { system_zone[length]=0; const char *root=strstr(system_zone,"/zoneinfo/"); if(root!=NULL) zone=root+10; }
 }
#endif
 bool utc=zone!=NULL && (strcmp(zone,"UTC")==0 || strcmp(zone,"Etc/UTC")==0);
 bool denver=zone!=NULL && strcmp(zone,"America/Denver")==0;
 if (!utc && !denver) return date_string_refuse("Date.toString: exact ICU long zone name unavailable for runtime TZ");
 int64_t seconds=(int64_t)floor(time/1000);
 int parts[7]; date_string_parts(seconds,parts);
 int offset=0;
 const char *name="Coordinated Universal Time";
 if (denver) {
#ifdef ADAMIC_TARGET_WASI
  return date_string_refuse("Date.toString: America/Denver requires a system zone database unavailable on wasm32-wasi");
#else
  // Keep libc and ICU within their shared Unix transition range. Wider dates
  // need V8's equivalent-year rules and are refused rather than approximated.
  if (seconds<0 || seconds>2147483647) return date_string_refuse("Date.toString: America/Denver outside the supported 1970-2038 zone transition range");
  time_t instant=(time_t)seconds;
  struct tm local;
  tzset();
  if (localtime_r(&instant,&local)==NULL) return date_string_refuse("Date.toString: localtime unavailable for America/Denver");
  // Denver's modern zone database has exactly these two offsets. Verify the
  // local calendar against UTC shifted by that offset before choosing its name.
  offset=local.tm_isdst>0 ? -360 : -420;
  date_string_parts(seconds+(int64_t)offset*60,parts);
  if (local.tm_isdst<0 || local.tm_year+1900!=parts[0] || local.tm_mon!=parts[1] || local.tm_mday!=parts[2] || local.tm_hour!=parts[3] || local.tm_min!=parts[4] || local.tm_sec!=parts[5]) return date_string_refuse("Date.toString: system America/Denver rules differ from the supported ICU zone");
  name=local.tm_isdst>0 ? "Mountain Daylight Time" : "Mountain Standard Time";
#endif
 }
 static const char *const days[]={"Sun","Mon","Tue","Wed","Thu","Fri","Sat"};
 static const char *const months[]={"Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"};
 char year[16],text[160];
 if(parts[0]<0) snprintf(year,sizeof year,"-%04d",-parts[0]);
 else snprintf(year,sizeof year,"%04d",parts[0]);
 int magnitude=offset<0?-offset:offset;
 int length=snprintf(text,sizeof text,"%s %s %02d %s %02d:%02d:%02d GMT%c%02d%02d (%s)",days[parts[6]],months[parts[1]],parts[2],year,parts[3],parts[4],parts[5],offset<0?'-':'+',magnitude/60,magnitude%60,name);
 if(length<0 || (size_t)length>=sizeof text) return date_string_refuse("Date.toString: formatting failed");
 return date_string_text(text);
}
