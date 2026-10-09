// UTC Date operations. Date internal slots are not own properties.
#ifndef ADAMIC_DATE_H
#define ADAMIC_DATE_H
adamic_object *adamic_date_new(double);
double adamic_date_value(const adamic_object *);
double adamic_date_now(void);
adamic_closure *adamic_date_now_function(void);
double adamic_date_utc(size_t, const double *);
double adamic_date_get(const adamic_object *, int);
double adamic_date_parse_iso(const adamic_string *);
adamic_string *adamic_date_iso(const adamic_object *);
adamic_string *adamic_date_utc_string(const adamic_object *);
bool adamic_date_has_own(const adamic_string *, bool);
#endif
