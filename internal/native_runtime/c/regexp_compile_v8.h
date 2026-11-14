#ifndef ADAMIC_REGEXP_COMPILE_V8_H
#define ADAMIC_REGEXP_COMPILE_V8_H
#include "regexp_compile_bytecode.h"
/* Raw bytecode remains available for reference identity. Runtime construction
 * must use this entry point, which applies the same compatibility refusals. */
adamic_regex_program *adamic_regex_compile_checked(adamic_regex_parse_result *result);
bool adamic_regex_compile_v8_check(adamic_regex_parse_result *result);
#endif
