#ifndef ADAMIC_TSGO_RUNTIME_H
#define ADAMIC_TSGO_RUNTIME_H
#include "adamic.h"
double adamic_tsgo_program(const adamic_string *config, const adamic_array *files);
adamic_object *adamic_tsgo_query(double handle, const adamic_string *file, double position);
adamic_object *adamic_tsgo_query_in(adamic_region *region, double handle, const adamic_string *file, double position);
adamic_string *adamic_tsgo_type_parts(double handle, const adamic_string *file, double start, double end, const adamic_string *kind);
adamic_string *adamic_tsgo_inspect(double handle, const adamic_string *file, double start, double end, const adamic_string *kind, const adamic_string *question);
void adamic_tsgo_release(double handle);
#endif
