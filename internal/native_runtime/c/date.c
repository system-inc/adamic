// Date text from its clipped epoch milliseconds, independent of libc time_t and TZ.
#include "adamic.h"
#include <stdint.h>
#include <stdio.h>

adamic_string *adamic_date_to_iso_string(double milliseconds) {
    if (!isfinite(milliseconds) || fabs(milliseconds) > 8640000000000000.) {
        const char message[] = "Invalid time value";
        adamic_throw_kind("RangeError", sizeof "RangeError" - 1, message, sizeof message - 1);
        return NULL;
    }
    // Date's TimeClip already made the value integral. Division is floored, not
    // truncated: -1 ms belongs to the preceding day, with a .999 fraction.
    int64_t ticks = (int64_t)milliseconds;
    int64_t days = ticks / 86400000;
    int64_t within = ticks % 86400000;
    if (within < 0) { days--; within += 86400000; }

    // Proleptic Gregorian civil date: 400-year eras starting on March 1.
    // The shifted epoch is 0000-03-01; floor the era for BCE dates too.
    int64_t shifted = days + 719468;
    int64_t era = (shifted >= 0 ? shifted : shifted - 146096) / 146097;
    int64_t day_of_era = shifted - era * 146097;
    int64_t year_of_era = (day_of_era - day_of_era / 1460 + day_of_era / 36524 - day_of_era / 146096) / 365;
    int year = (int)(year_of_era + era * 400);
    int64_t day_of_year = day_of_era - (365 * year_of_era + year_of_era / 4 - year_of_era / 100);
    int month_part = (int)((5 * day_of_year + 2) / 153);
    int day = (int)(day_of_year - (153 * month_part + 2) / 5 + 1);
    int month = month_part + (month_part < 10 ? 3 : -9);
    year += month <= 2;
    int hour = (int)(within / 3600000);
    int minute = (int)(within / 60000 % 60);
    int second = (int)(within / 1000 % 60);
    int fraction = (int)(within % 1000);
    char buffer[32];
    int length;
    if (year >= 0 && year <= 9999) {
        length = snprintf(buffer, sizeof buffer, "%04d-%02d-%02dT%02d:%02d:%02d.%03dZ", year, month, day, hour, minute, second, fraction);
    } else {
        length = snprintf(buffer, sizeof buffer, "%c%06d-%02d-%02dT%02d:%02d:%02d.%03dZ", year < 0 ? '-' : '+', year < 0 ? -year : year, month, day, hour, minute, second, fraction);
    }
    return adamic_decode_utf8((const unsigned char *)buffer, (size_t)length);
}
