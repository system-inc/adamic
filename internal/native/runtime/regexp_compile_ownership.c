#include "adamic.h"
#ifdef ADAMIC_REGEXP_RUNTIME_COMPILER
#include <string.h>
/* Private counted storage keeps dynamic instructions and group shapes alive. */
static const char *const regex_owned_names[] = {"__program","lastIndex","source","flags","global","ignoreCase","multiline","unicode","sticky","hasIndices","unicodeSets","dotAll","#compiler"};
static const bool regex_owned_references[] = {false,false,true,true,false,false,false,false,false,false,false,false,true};
static const adamic_shape regex_owned_shape = {13,regex_owned_names,regex_owned_references,NULL,NULL};
adamic_object *adamic_regex_new_owned(const adamic_regex_program *program,adamic_string *source,adamic_string *flags,adamic_array *storage) {
    adamic_object *plain=adamic_regex_new(program,source,flags);
    adamic_object *result=adamic_object_new(&regex_owned_shape);
    memcpy(result->slots,plain->slots,12*sizeof(*result->slots));
    result->slots[2].reference=adamic_retain(source);result->slots[3].reference=adamic_retain(flags);
    result->slots[12].reference=adamic_retain(storage);adamic_release(plain);return result;
}
#endif
