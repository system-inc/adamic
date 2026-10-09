// input.c: what a program reads from outside, opened in 0.2, its arguments and files, and the one
// way it writes back besides its output: writeTextFile.
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
#ifdef ADAMIC_TARGET_WASI
#include <sys/stat.h>
#endif

// new_string makes a string of length bytes, references 1, its bytes right after it in one block, the
// same layout string.c makes, so the heap frees it the same way.
static adamic_string *new_string(size_t length) {
	adamic_string *string = adamic_allocate(sizeof *string + length, adamic_kind_string);
	string->length = length;
	string->bytes = (const char *)(string + 1);
	string->units = 0;
	string->index = NULL;
	string->owner = NULL;
	string->capacity = length;
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

// ascii_prefix scans only complete words inside the input. memcpy permits unaligned
// pointers without aliasing violations; the repeated high-bit mask is endian independent.
static size_t ascii_prefix(const unsigned char *bytes, size_t length) {
	size_t offset = 0;
	while (length - offset >= sizeof(uint64_t)) {
		uint64_t word;
		memcpy(&word, bytes + offset, sizeof word);
		if ((word & UINT64_C(0x8080808080808080)) != 0) {
			break;
		}
		offset += sizeof word;
	}
	while (offset < length && bytes[offset] < 0x80) {
		offset++;
	}
	return offset;
}

// decode makes a string of bytes decoded as WHATWG UTF-8, which the caller owns. What it makes is
// always valid UTF-8, so no lone surrogate can come in from outside.
static adamic_string *decode(const unsigned char *bytes, size_t length) {
	size_t ascii = ascii_prefix(bytes, length);
	size_t size = ascii;
	unsigned point;
	for (size_t offset = ascii; offset < length;) {
		offset += decode_step(bytes, length, offset, &point);
		size += encoded_size(point);
	}
	adamic_string *string = new_string(size);
	if (ascii == length) {
		string->units = length + 1;
	}
	unsigned char *cursor = (unsigned char *)string->bytes;
	if (ascii != 0) {
		memcpy(cursor, bytes, ascii);
		cursor += ascii;
	}
	for (size_t offset = ascii; offset < length;) {
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

adamic_string *adamic_decode_utf8(const unsigned char *bytes, size_t length) {
	return decode(bytes, length);
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
static const adamic_shape ok_shape = {2, ok_names, both_references, NULL, NULL};
static const adamic_shape error_shape = {2, error_names, both_references, NULL, NULL};
static adamic_string ok_kind = ADAMIC_STRING("Ok");
static adamic_string error_kind = ADAMIC_STRING("Error");

// failure is readTextFile's and writeTextFile's { kind: 'Error', message }, the message in Adamic's
// own words, the same on every platform: "cannot read <path>: <reason>", or "cannot write". The
// JavaScript runtime says the same (oracle/adamic.mjs), from Node's error codes, which are these
// errno names. What's missing when a write can't find its path is a directory, not the file.
static adamic_object *failure(const adamic_string *path, int error, bool writing) {
	static adamic_string read_prefix = ADAMIC_STRING("cannot read ");
	static adamic_string write_prefix = ADAMIC_STRING("cannot write ");
	static adamic_string missing_file = ADAMIC_STRING(": no such file");
	static adamic_string missing_directory = ADAMIC_STRING(": no such directory");
	static adamic_string denied = ADAMIC_STRING(": permission denied");
	static adamic_string directory = ADAMIC_STRING(": is a directory");
	static adamic_string failed = ADAMIC_STRING(": failed");
	adamic_string *reason = &failed;
	switch (error) {
	case ENOENT:
	case ENOTDIR:
		reason = writing ? &missing_directory : &missing_file;
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
	result->slots[1].reference = adamic_string_concat(3, (adamic_string *const[]){writing ? &write_prefix : &read_prefix, (adamic_string *)path, reason});
	return result;
}

// utf8 is a string's bytes as Node's encoding to UTF-8 makes them, terminated, in a buffer the
// caller frees: a lone surrogate, which a string holds as WTF-8, becomes U+FFFD's bytes, the same
// length.
static char *utf8(const adamic_string *string) {
	char *bytes = malloc(string->length + 1);
	if (bytes == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	memcpy(bytes, string->bytes, string->length);
	for (size_t at = 0; at + 3 <= string->length; at++) {
		if ((unsigned char)bytes[at] == 0xed && (unsigned char)bytes[at + 1] >= 0xa0) {
			memcpy(bytes + at, "\xef\xbf\xbd", 3);
			at += 2;
		}
	}
	bytes[string->length] = '\0';
	return bytes;
}

// file_name is a path as the bytes Node would name the file by. A path holding a NUL can name no
// file, and Node refuses it before asking, so that's NULL here.
static char *file_name(const adamic_string *path) {
	if (memchr(path->bytes, 0, path->length) != NULL) {
		return NULL;
	}
	return utf8(path);
}

// read_all reads every byte from a descriptor into a buffer the caller frees, through partial reads
// and interrupted calls. It returns 0, or the errno that stopped it.
static int read_all(int descriptor, unsigned char **bytes, size_t *length) {
	// Both outputs are defined on every path, so a caller that reads them after an error reads
	// nothing it didn't set.
	*bytes = NULL;
	*length = 0;
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
			// read sets errno when it fails; an error with none is still an error, never success.
			int error = errno != 0 ? errno : EIO;
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
	// What was printed comes first, as on Node: the file may be stdin, waiting on a prompt.
	adamic_output_flush();
#ifdef ADAMIC_TARGET_WASI
	// WASI libc cannot resolve an empty capability path. Node treats it as ENOENT.
	if (path->length == 0) {
		return failure(path, ENOENT, false);
	}
#endif
	char *name = file_name(path);
	if (name == NULL) {
		return failure(path, 0, false);
	}
	int descriptor;
	do {
		descriptor = open(name, O_RDONLY | O_CLOEXEC);
	} while (descriptor < 0 && errno == EINTR);
	free(name);
	if (descriptor < 0) {
		return failure(path, errno, false);
	}
#ifdef ADAMIC_TARGET_WASI
	// fd_read on a directory returns EBADF in WASI; Node reports EISDIR instead.
	struct stat status;
	if (fstat(descriptor, &status) == 0 && S_ISDIR(status.st_mode)) {
		close(descriptor);
		return failure(path, EISDIR, false);
	}
#endif
	unsigned char *bytes;
	size_t length;
	int error = read_all(descriptor, &bytes, &length);
	close(descriptor);
	if (error != 0) {
		return failure(path, error, false);
	}
	// Node's readFileSync throws ERR_STRING_TOO_LONG for a file of V8's longest string in bytes or
	// more, whatever they decode to (536,870,888 bytes of é, half as many units, fails too, and
	// 536,870,887 bytes reads), and readTextFile answers that as a failure. So here too.
	if (length >= ADAMIC_STRING_MAX_UNITS) {
		free(bytes);
		return failure(path, 0, false);
	}
	adamic_string *text = decode(bytes, length);
	free(bytes);
	adamic_object *result = adamic_object_new(&ok_shape);
	result->slots[0].reference = &ok_kind;
	result->slots[1].reference = text;
	return result;
}

// The shape writeTextFile's success comes in.
static const char *const written_names[] = {"kind"};
static const adamic_shape written_shape = {1, written_names, both_references, NULL, NULL};

// write_all writes every byte, through partial writes and interrupted calls, and returns 0 or the
// errno that stopped it.
static int write_all(int descriptor, const char *bytes, size_t length) {
	while (length > 0) {
		ssize_t written = write(descriptor, bytes, length);
		if (written < 0) {
			if (errno == EINTR) {
				continue;
			}
			return errno;
		}
		bytes += written;
		length -= (size_t)written;
	}
	return 0;
}

// adamic_write_text_file opens as Node's writeFileSync does (flag 'w', mode 0666 before the umask):
// the file made if it isn't there, emptied if it is, and then the text written whole.
adamic_object *adamic_write_text_file(const adamic_string *path, const adamic_string *text) {
	// What was printed comes first, as on Node: the file may be stdout or stderr themselves.
	adamic_output_flush();
#ifdef ADAMIC_TARGET_WASI
	if (path->length == 0) {
		return failure(path, ENOENT, true);
	}
#endif
	char *name = file_name(path);
	if (name == NULL) {
		return failure(path, 0, true);
	}
	int descriptor = -1;
	bool borrowed = false;
#ifdef ADAMIC_TARGET_WASI
	// Preview 1 supplies stdout/stderr as descriptors, but no /dev namespace.
	// Node's WASI host reports a pipe as a stream socket. Reopening these stream
	// aliases on the host fails; write the supplied capability without closing it.
	// Regular files still go through open, preserving Node's truncation semantics.
	int stream = strcmp(name, "/dev/stdout") == 0 ? STDOUT_FILENO :
		strcmp(name, "/dev/stderr") == 0 ? STDERR_FILENO : -1;
	struct stat status;
	if (stream >= 0 && fstat(stream, &status) == 0 &&
		(S_ISFIFO(status.st_mode) || S_ISCHR(status.st_mode) || S_ISSOCK(status.st_mode))) {
		descriptor = stream;
		borrowed = true;
	}
#endif
	if (!borrowed) {
		do {
			descriptor = open(name, O_WRONLY | O_CREAT | O_TRUNC | O_CLOEXEC, 0666);
		} while (descriptor < 0 && errno == EINTR);
	}
	free(name);
	if (descriptor < 0) {
		return failure(path, errno, true);
	}
	char *bytes = utf8(text);
	int error = write_all(descriptor, bytes, text->length);
	free(bytes);
	if (!borrowed && close(descriptor) != 0 && error == 0 && errno != EINTR) {
		error = errno;
	}
	if (error != 0) {
		return failure(path, error, true);
	}
	adamic_object *result = adamic_object_new(&written_shape);
	result->slots[0].reference = &ok_kind;
	return result;
}

char *adamic_path_bytes(const adamic_string *path) {
	return file_name(path);
}
