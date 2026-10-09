// Restricted V8 13.6.233 library error construction: src/builtins/builtins-number.cc,
// string-repeat.tq, string-pad.tq and src/objects/objects.cc. Message templates are
// from src/common/message-template.h. BSD-3-Clause in THIRD_PARTY_NOTICES.md.
#include "library_errors.h"
#include <string.h>

void adamic_library_throw(const char *name, const char *message, size_t length) {
 adamic_string *text=adamic_string_allocate(length);
 memcpy((char *)text->bytes,message,length);
 adamic_object *error=adamic_error_new(text);
 adamic_release(text);
 adamic_string *type=adamic_string_allocate(strlen(name));
 memcpy((char *)type->bytes,name,type->length);
 adamic_release(error->slots[0].reference);
 error->slots[0].reference=type;
 adamic_error_tag(error);
 adamic_thrown=error;
}
bool adamic_library_concat_check(size_t count, adamic_string *const parts[]) {
 double units=0;
 for(size_t i=0;i<count;i++) units+=adamic_string_length(parts[i]);
 if(units>ADAMIC_STRING_MAX_UNITS) {
  const char message[]="Invalid string length";
  adamic_library_throw("RangeError",message,sizeof message-1);
  return false;
 }
 return true;
}
adamic_string *adamic_library_concat(size_t count, adamic_string *const parts[]) {
 return adamic_library_concat_check(count,parts) ? adamic_string_concat(count,parts) : NULL;
}

// Test nominal identity, not the writable Error.name property.
bool adamic_library_error_is_type(const adamic_object *error, const adamic_string *name) {
 unsigned definition = name->length == 9 && memcmp(name->bytes,"TypeError",9) == 0 ? (1u<<30)+1 : (1u<<30)+2;
 for (const adamic_class *class=error->class;class!=NULL;class=class->base) {
  if (class->definition==definition) { return true; }
 }
 return false;
}
