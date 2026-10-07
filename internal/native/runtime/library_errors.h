// A library error that has no compiler-generated unwinding guard ends as an ordinary
// uncaught JavaScript error. Guards in lowering make the supported failures catchable.
#ifndef ADAMIC_LIBRARY_ERRORS_H
#define ADAMIC_LIBRARY_ERRORS_H
#include <stddef.h>
_Noreturn void adamic_uncaught_library_error(const char *message, size_t length);
#endif
