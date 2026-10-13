#ifndef ADAMIC_REGEXP_COMPILE_BYTECODE_H
#define ADAMIC_REGEXP_COMPILE_BYTECODE_H
#include "regexp_compile_sets.h"
/* The VM representation and opcode order are the existing matcher's. This
 * program and every referenced array belong to the parse-result arena. */
adamic_regex_program *adamic_regex_compile_bytecode(adamic_regex_parse_result *result);
#endif
