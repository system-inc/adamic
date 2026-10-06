#ifndef ADAMIC_TSGO_RUNTIME_H
#define ADAMIC_TSGO_RUNTIME_H
#include "adamic.h"
double adamic_tsgo_program(const adamic_string *config, const adamic_array *files);
adamic_object *adamic_tsgo_query(double handle, const adamic_string *file, double position);
adamic_object *adamic_tsgo_query_in(adamic_region *region, double handle, const adamic_string *file, double position);
void adamic_tsgo_release(double handle);
#endif
