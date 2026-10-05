// adamic.h: the Adamic runtime. Small, plain C11, compiled with every program.

#ifndef ADAMIC_H
#define ADAMIC_H

#include <stddef.h>

enum adamic_stream {
	adamic_stdout = 1,
	adamic_stderr = 2,
};

// adamic_write_line writes length bytes and a newline, as console.log does with one string.
void adamic_write_line(enum adamic_stream stream, const char *bytes, size_t length);

// ADAMIC_NUMBER_FORMAT_MAX holds the longest number text, "-1.2345678901234567e-308", with room.
#define ADAMIC_NUMBER_FORMAT_MAX 32

// adamic_number_format writes value as JavaScript's String(value) does and returns the length. The
// buffer is not terminated.
size_t adamic_number_format(double value, char buffer[ADAMIC_NUMBER_FORMAT_MAX]);

// adamic_panic writes "adamic: panic: <message>" to stderr and exits 70 (EX_SOFTWARE).
_Noreturn void adamic_panic(const char *message, size_t length);

#endif
