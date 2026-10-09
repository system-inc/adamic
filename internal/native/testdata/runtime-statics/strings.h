#include "harness.h"
static adamic_string *text;
static void prepare(void) {
 text = adamic_string_allocate(STRING_BYTES * 128);
 for (size_t index = 0; index < 128; index++) { memcpy((char *)text->bytes + STRING_BYTES * index, STRING_TEXT, STRING_BYTES); }
 adamic_share(text);
}
static void cleanup(void) { adamic_release(text); }
static double exercise(size_t index) {
 size_t length = (size_t)adamic_string_length(text);
 if (length != STRING_UNITS * 128) { abort(); }
 double unit = adamic_string_char_code_at(text, (double)(index % length));
 adamic_string *slice = adamic_string_slice(text, 1, (double)length - 1, true);
 if (adamic_string_length(slice) != (double)length - 2) { abort(); }
 adamic_release(slice);
 return unit + 1;
}
