#ifndef ADAMIC_VIEW_NOMINAL_GRAPH_H
#define ADAMIC_VIEW_NOMINAL_GRAPH_H
#include "adamic.h"

typedef struct {
 const char *name;
 unsigned int schema;
 bool optional;
} adamic_nominal_graph_field;
typedef struct {
 const char *name;
 const adamic_class *nominal;
 bool null_allowed;
 bool undefined_allowed;
 bool array;
 unsigned int element;
 const adamic_nominal_graph_field *fields;
 size_t field_count;
} adamic_nominal_graph_schema;
void adamic_nominal_graph_check(const void *value,unsigned int root,const adamic_nominal_graph_schema *schemas,const char *where);
#endif
