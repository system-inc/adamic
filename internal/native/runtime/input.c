// input.c: what a program reads from outside, opened in 0.2: its arguments, and files.
//
// Both arrive as bytes, and a program sees strings, so both are decoded exactly as Node decodes them:
// UTF-8 by the WHATWG decoder's rules, every invalid sequence one U+FFFD, and a byte-order mark kept
// as U+FEFF. The oracle holds that to Node with a sweep over malformed sequences.

// O_CLOEXEC is POSIX.1-2008's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"

#include <errno.h>
#include <fcntl.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// new_string makes a string of length bytes, references 1, its bytes right after it in one block, the
// same layout string.c makes, so the heap frees it the same way.
static adamic_string *new_string(size_t length) {
	adamic_string *string = adamic_allocate(sizeof *string + length, adamic_kind_string);
	string->length = length;
	string->bytes = (const char *)(string + 1);
	return string;
}

// decode_step reads one code point of WHATWG UTF-8 at offset, and returns how many bytes it took.
// An invalid sequence is U+FFFD and takes only the bytes before the one that broke it, so that byte
// is read again as the start of what follows; a sequence cut off by the end is one U+FFFD.
static size_t decode_step(const unsigned char *bytes, size_t length, size_t offset, unsigned *point) {
	unsigned char lead = bytes[offset];
	if (lead < 0x80) {
		*point = lead;
		return 1;
	}
	unsigned char lower = 0x80, upper = 0xbf;
	size_t needed;
	unsigned value;
	if (lead >= 0xc2 && lead <= 0xdf) {
		needed = 1;
		value = lead & 0x1f;
	} else if (lead >= 0xe0 && lead <= 0xef) {
		// E0 can't begin an overlong form, and ED can't begin a surrogate.
		if (lead == 0xe0) {
			lower = 0xa0;
		}
		if (lead == 0xed) {
			upper = 0x9f;
		}
		needed = 2;
		value = lead & 0x0f;
	} else if (lead >= 0xf0 && lead <= 0xf4) {
		// F0 can't begin an overlong form, and F4 can't go past U+10FFFF.
		if (lead == 0xf0) {
			lower = 0x90;
		}
		if (lead == 0xf4) {
			upper = 0x8f;
		}
		needed = 3;
		value = lead & 0x07;
	} else {
		*point = 0xfffd;
		return 1;
	}
	size_t seen = 1;
	while (seen <= needed) {
		if (offset + seen >= length) {
			*point = 0xfffd;
			return seen;
		}
		unsigned char byte = bytes[offset + seen];
		if (byte < lower || byte > upper) {
			*point = 0xfffd;
			return seen;
		}
		lower = 0x80;
		upper = 0xbf;
		value = (value << 6) | (byte & 0x3f);
		seen++;
	}
	*point = value;
	return seen;
}

// encoded_size is how many bytes UTF-8 writes a code point in.
static size_t encoded_size(unsigned point) {
	return point < 0x80 ? 1 : point < 0x800 ? 2 : point < 0x10000 ? 3 : 4;
}

// decode makes a string of bytes decoded as WHATWG UTF-8, which the caller owns. What it makes is
// always valid UTF-8, so no lone surrogate can come in from outside.
static adamic_string *decode(const unsigned char *bytes, size_t length) {
	size_t size = 0;
	unsigned point;
	for (size_t offset = 0; offset < length;) {
		offset += decode_step(bytes, length, offset, &point);
		size += encoded_size(point);
	}
	adamic_string *string = new_string(size);
	unsigned char *cursor = (unsigned char *)string->bytes;
	for (size_t offset = 0; offset < length;) {
		offset += decode_step(bytes, length, offset, &point);
		switch (encoded_size(point)) {
		case 1:
			*cursor++ = (unsigned char)point;
			break;
		case 2:
			*cursor++ = (unsigned char)(0xc0 | (point >> 6));
			*cursor++ = (unsigned char)(0x80 | (point & 0x3f));
			break;
		case 3:
			*cursor++ = (unsigned char)(0xe0 | (point >> 12));
			*cursor++ = (unsigned char)(0x80 | ((point >> 6) & 0x3f));
			*cursor++ = (unsigned char)(0x80 | (point & 0x3f));
			break;
		default:
			*cursor++ = (unsigned char)(0xf0 | (point >> 18));
			*cursor++ = (unsigned char)(0x80 | ((point >> 12) & 0x3f));
			*cursor++ = (unsigned char)(0x80 | ((point >> 6) & 0x3f));
			*cursor++ = (unsigned char)(0x80 | (point & 0x3f));
			break;
		}
	}
	return string;
}

static int argument_count;
static char **argument_values;

void adamic_arguments_save(int count, char **values) {
	argument_count = count;
	argument_values = values;
}

adamic_array *adamic_program_arguments(void) {
	size_t count = argument_count > 1 ? (size_t)argument_count - 1 : 0;
	adamic_array *arguments = adamic_array_new(count, true);
	for (size_t index = 0; index < count; index++) {
		const char *argument = argument_values[index + 1];
		adamic_array_push(arguments, (adamic_value){.reference = decode((const unsigned char *)argument, strlen(argument))});
	}
	return arguments;
}

// The two shapes readTextFile's result comes in, and the kinds that tell them apart.
static const char *const ok_names[] = {"kind", "text"};
static const char *const error_names[] = {"kind", "message"};
static const bool both_references[] = {true, true};
static const adamic_shape ok_shape = {2, ok_names, both_references};
static const adamic_shape error_shape = {2, error_names, both_references};
static adamic_string ok_kind = ADAMIC_STRING("Ok");
static adamic_string error_kind = ADAMIC_STRING("Error");

// failure is readTextFile's { kind: 'Error', message }, the message in Adamic's own words, the same
// on every platform: "cannot read <path>: <reason>". The JavaScript runtime says the same
// (oracle/adamic.mjs), from Node's error codes, which are these errno names.
static adamic_object *failure(const adamic_string *path, int error) {
	static adamic_string prefix = ADAMIC_STRING("cannot read ");
	static adamic_string missing = ADAMIC_STRING(": no such file");
	static adamic_string denied = ADAMIC_STRING(": permission denied");
	static adamic_string directory = ADAMIC_STRING(": is a directory");
	static adamic_string failed = ADAMIC_STRING(": failed");
	adamic_string *reason = &failed;
	switch (error) {
	case ENOENT:
	case ENOTDIR:
		reason = &missing;
		break;
	case EACCES:
	case EPERM:
		reason = &denied;
		break;
	case EISDIR:
		reason = &directory;
		break;
	}
	adamic_object *result = adamic_object_new(&error_shape);
	result->slots[0].reference = &error_kind;
	result->slots[1].reference = adamic_string_concat(3, (adamic_string *const[]){&prefix, (adamic_string *)path, reason});
	return result;
}

// file_name is a path as the bytes Node would name the file by, terminated: a lone surrogate, which
// a string holds as WTF-8, becomes U+FFFD's bytes, as Node's own encoding to UTF-8 makes it. A path
// holding a NUL can name no file, and Node refuses it before asking, so that's NULL here.
static char *file_name(const adamic_string *path) {
	if (memchr(path->bytes, 0, path->length) != NULL) {
		return NULL;
	}
	char *name = malloc(path->length + 1);
	if (name == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	memcpy(name, path->bytes, path->length);
	for (size_t at = 0; at + 3 <= path->length; at++) {
		if ((unsigned char)name[at] == 0xed && (unsigned char)name[at + 1] >= 0xa0) {
			memcpy(name + at, "\xef\xbf\xbd", 3);
			at += 2;
		}
	}
	name[path->length] = '\0';
	return name;
}

// read_all reads every byte from a descriptor into a buffer the caller frees, through partial reads
// and interrupted calls. It returns 0, or the errno that stopped it.
static int read_all(int descriptor, unsigned char **bytes, size_t *length) {
	size_t capacity = 4096, used = 0;
	unsigned char *buffer = malloc(capacity);
	for (;;) {
		if (buffer == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		if (used == capacity) {
			capacity *= 2;
			unsigned char *grown = realloc(buffer, capacity);
			if (grown == NULL) {
				free(buffer);
			}
			buffer = grown;
			continue;
		}
		ssize_t got = read(descriptor, buffer + used, capacity - used);
		if (got < 0) {
			if (errno == EINTR) {
				continue;
			}
			int error = errno;
			free(buffer);
			return error;
		}
		if (got == 0) {
			*bytes = buffer;
			*length = used;
			return 0;
		}
		used += (size_t)got;
	}
}

adamic_object *adamic_read_text_file(const adamic_string *path) {
	char *name = file_name(path);
	if (name == NULL) {
		return failure(path, 0);
	}
	int descriptor;
	do {
		descriptor = open(name, O_RDONLY | O_CLOEXEC);
	} while (descriptor < 0 && errno == EINTR);
	free(name);
	if (descriptor < 0) {
		return failure(path, errno);
	}
	unsigned char *bytes;
	size_t length;
	int error = read_all(descriptor, &bytes, &length);
	close(descriptor);
	if (error != 0) {
		return failure(path, error);
	}
	adamic_object *result = adamic_object_new(&ok_shape);
	result->slots[0].reference = &ok_kind;
	result->slots[1].reference = decode(bytes, length);
	free(bytes);
	return result;
}
