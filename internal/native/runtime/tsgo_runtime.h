#ifndef ADAMIC_TSGO_RUNTIME_H
#define ADAMIC_TSGO_RUNTIME_H
#include "adamic.h"
adamic_object *adamic_tsgo_program(const adamic_string *config, const adamic_array *files);
adamic_object *adamic_tsgo_program_in(adamic_region *region, const adamic_string *config, const adamic_array *files);
adamic_object *adamic_tsgo_type_parts(double handle, const adamic_string *file, double start, double end, const adamic_string *kind);
adamic_object *adamic_tsgo_type_parts_in(adamic_region *region, double handle, const adamic_string *file, double start, double end, const adamic_string *kind);
adamic_object *adamic_tsgo_inspect(double handle, const adamic_string *file, double start, double end, const adamic_string *kind, const adamic_string *question);
adamic_object *adamic_tsgo_inspect_in(adamic_region *region, double handle, const adamic_string *file, double start, double end, const adamic_string *kind, const adamic_string *question);
adamic_object *adamic_tsgo_release(double handle);
adamic_object *adamic_tsgo_release_in(adamic_region *region, double handle);
adamic_object *adamic_tsgo_query(double handle, const adamic_string *file, double position);
adamic_object *adamic_tsgo_query_in(adamic_region *region, double handle, const adamic_string *file, double position);
#endif
