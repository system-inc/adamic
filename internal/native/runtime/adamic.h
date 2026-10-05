// adamic.h: the Adamic runtime. Small, plain C11, compiled with every program.

#ifndef ADAMIC_H
#define ADAMIC_H

#include <stddef.h>

enum adamic_stream {
	adamic_stdout = 1,
	adamic_stderr = 2,
};

// adamic_string is an immutable string: UTF-8 bytes, reference counted. references 0 is a constant
// the program spelled out, immortal; a built string starts at 1 (string.c).
typedef struct adamic_string {
	size_t references;
	size_t length;
	const char *bytes;
} adamic_string;

// ADAMIC_STRING is a constant: ADAMIC_STRING("text") as a static adamic_string's initializer.
#define ADAMIC_STRING(text) {0, sizeof text - 1, text}

adamic_string *adamic_retain(adamic_string *string);
void adamic_release(adamic_string *string);

// These return a reference the caller owns.
adamic_string *adamic_string_from_number(double value);
adamic_string *adamic_string_concat(size_t count, adamic_string *const parts[]);

// The strings every program has: "", and String(true) and String(false).
extern adamic_string adamic_string_empty;
extern adamic_string adamic_string_true;
extern adamic_string adamic_string_false;

// adamic_string_equal is ===.
int adamic_string_equal(const adamic_string *left, const adamic_string *right);

// adamic_write_line writes a string and a newline, as console.log does with one string.
void adamic_write_line(enum adamic_stream stream, const adamic_string *string);

// ADAMIC_NUMBER_FORMAT_MAX holds the longest number text, "-1.2345678901234567e-308", with room.
#define ADAMIC_NUMBER_FORMAT_MAX 32

// adamic_number_format writes value as JavaScript's String(value) does and returns the length. The
// buffer is not terminated.
size_t adamic_number_format(double value, char buffer[ADAMIC_NUMBER_FORMAT_MAX]);

// adamic_power is JavaScript's ** (number.c).
double adamic_power(double base, double exponent);

// adamic_panic writes "adamic: panic: <message>" to stderr and exits 70 (EX_SOFTWARE).
_Noreturn void adamic_panic(const char *message, size_t length);

// adamic_unreachable ends a function the checker proved always returns. Reaching it is a compiler
// bug, and it says so rather than returning garbage.
_Noreturn void adamic_unreachable(void);

#endif
