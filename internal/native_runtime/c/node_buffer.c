// Node v24.19.0's Buffer surface. UTF-8 below ports V8's Utf8DfaDecoder
// and Utf8::ValueOfIncremental: a rejected continuation is reprocessed, and an
// incomplete sequence at EOF produces exactly one replacement character.
// Copyright 2007-2010 the V8 project authors. BSD-3-Clause;
// THIRD_PARTY_NOTICES.md. Base64 and hex port Node's nbytes fallback scalar
// loops (MIT, same notices).
#include "adamic.h"
#include <stdlib.h>
#include <string.h>

static const uint8_t transitions[] = {
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 00-0F
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 10-1F
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 20-2F
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 30-3F
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 40-4F
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 50-5F
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 60-6F
	0,	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 70-7F
	1,	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 80-8F
	2,	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, // 90-9F
	3,	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, // A0-AF
	3,	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, // B0-BF
	9,	9, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, // C0-CF
	4,	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, // D0-DF
	10, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6, 5, 5, // E0-EF
	11, 7, 7, 7, 8, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, // F0-FF
};
static const uint8_t states[] = {
	0,	0,	0,	0,	0,	0,	0,	0,	0,	0, 0,  0,  // REJECT = 0
	12, 0,	0,	0,	24, 36, 48, 60, 72, 0, 84, 96, // ACCEPT = 12
	0,	12, 12, 12, 0,	0,	0,	0,	0,	0, 0,  0,  // 2-byte = 24
	0,	24, 24, 24, 0,	0,	0,	0,	0,	0, 0,  0,  // 3-byte = 36
	0,	24, 24, 0,	0,	0,	0,	0,	0,	0, 0,  0,  // 3-byte low/mid = 48
	0,	36, 36, 36, 0,	0,	0,	0,	0,	0, 0,  0,  // 4-byte = 60
	0,	36, 0,	0,	0,	0,	0,	0,	0,	0, 0,  0,  // 4-byte low = 72
	0,	0,	0,	24, 0,	0,	0,	0,	0,	0, 0,  0,  // 3-byte high = 84
	0,	0,	36, 36, 0,	0,	0,	0,	0,	0, 0,  0,  // 4-byte mid/high = 96
};

static size_t put_point(unsigned char *out, uint32_t point)
{
	if (point < 0x80) {
		out[0] = (unsigned char)point;
		return 1;
	}
	if (point < 0x800) {
		out[0] = (unsigned char)(0xc0 | (point >> 6));
		out[1] = (unsigned char)(0x80 | (point & 63));
		return 2;
	}
	if (point < 0x10000) {
		out[0] = (unsigned char)(0xe0 | (point >> 12));
		out[1] = (unsigned char)(0x80 | ((point >> 6) & 63));
		out[2] = (unsigned char)(0x80 | (point & 63));
		return 3;
	}
	out[0] = (unsigned char)(0xf0 | (point >> 18));
	out[1] = (unsigned char)(0x80 | ((point >> 12) & 63));
	out[2] = (unsigned char)(0x80 | ((point >> 6) & 63));
	out[3] = (unsigned char)(0x80 | (point & 63));
	return 4;
}
static adamic_string *decode_utf8(const unsigned char *bytes, size_t length)
{
	// Each input byte can produce at most three output bytes (a replacement).
	adamic_string *text = adamic_string_allocate(length * 3);
	unsigned char *out = (unsigned char *)text->bytes;
	size_t written = 0;
	uint8_t state = 12;
	uint32_t point = 0;
	for (size_t at = 0; at < length;) {
		uint8_t old = state, byte = bytes[at++], type = transitions[byte];
		state = states[state + type];
		point = (point << 6) | (byte & (0x7f >> (type >> 1)));
		if (state == 12) {
			written += put_point(out + written, point);
			point = 0;
		} else if (state == 0) {
			state = 12;
			point = 0;
			if (old != 12) {
				at--;
			}
			written += put_point(out + written, 0xfffd);
		}
	}
	if (state != 12) {
		written += put_point(out + written, 0xfffd);
	}
	text->length = written;
	return text;
}
// Buffer.from(aBuffer) copies the bytes the source sees. Buffer.from(number[])
// isn't here: the native emitter calls adamic_typed_array_from_numbers, whose
// ToUint8 is a Uint8Array's and Node's.
adamic_typed_array *adamic_node_buffer_copy(const adamic_typed_array *source)
{
	adamic_typed_array *buffer =
		adamic_typed_array_make(adamic_typed_array_uint8, source->length);
	if (source->length > 0) {
		memcpy(buffer->data, source->data, source->length);
	}
	return buffer;
}
double adamic_node_buffer_set(adamic_typed_array *buffer, double index,
							  double value)
{
	// A Buffer write outside the bounds, or at an index that isn't whole, has
	// no effect, as on Node: a compiler-checked typed write panics there, so a
	// Buffer's writes come here instead.
	if (index >= 0 && index < (double)buffer->length && index == trunc(index)) {
		adamic_typed_array_write_converted(buffer, (size_t)index, value);
	}
	return value;
}
// received writes a number as Node's ERR_OUT_OF_RANGE writes the one it was
// given, and returns its length: an integer past 2^32 in magnitude as String
// writes it with addNumericalSeparator's underscores, one before every three
// characters from the right past a sign, exponent notation included (1e+300
// is 1e+_300); any other as String writes it. Node writes the others with
// util.inspect, which differs from String only for -0, never out of range
// here. process.c's exit code text groups the same way; its mutants pin its
// text where it is, so this stays its own.
static size_t received(double value, char out[ADAMIC_NUMBER_FORMAT_MAX * 2])
{
	char number[ADAMIC_NUMBER_FORMAT_MAX];
	size_t length = adamic_number_format(value, number);
	bool grouped = isfinite(value) && value == trunc(value) &&
				   fabs(value) > 4294967296.0;
	size_t sign = number[0] == '-' ? 1 : 0, used = 0;
	for (size_t index = 0; index < length; index++) {
		if (grouped && index > sign && (length - index) % 3 == 0) {
			out[used++] = '_';
		}
		out[used++] = number[index];
	}
	return used;
}
// Buffer.allocUnsafe(size), as Node v24.19.0's: the size is checked before
// it's truncated, so -0.5 is out of range where a Uint8Array's ToIndex would
// make it 0, and only then allocated. Out of range, NaN and the infinities
// among it, is ERR_OUT_OF_RANGE; a size that can't be allocated, one past
// SIZE_MAX on wasm32 among them, is ERR_MEMORY_ALLOCATION_FAILED. Both are
// RangeErrors with a code, left pending, and no Buffer comes back; only the
// first's String brackets its code, as Node's does. Node's bytes are
// unspecified, a small Buffer's from a shared pool; these are zero.
adamic_typed_array *adamic_node_buffer_alloc_unsafe(double size)
{
	if (!(size >= 0) || size > 9007199254740991.0) {
		static const char prefix[] =
			"The value of \"size\" is out of range. It must be >= 0 && <= "
			"9007199254740991. Received ";
		char message[sizeof prefix + ADAMIC_NUMBER_FORMAT_MAX * 2];
		memcpy(message, prefix, sizeof prefix - 1);
		size_t written =
			sizeof prefix - 1 + received(size, message + sizeof prefix - 1);
		adamic_throw_coded("RangeError", "ERR_OUT_OF_RANGE", message, written, true);
		return NULL;
	}
	double whole = trunc(size);
	adamic_typed_array *buffer =
		whole > (double)SIZE_MAX ?
			NULL :
			adamic_typed_array_try_make(adamic_typed_array_uint8,
										(size_t)whole);
	if (buffer == NULL) {
		static const char message[] = "Failed to allocate memory";
		adamic_throw_coded("RangeError", "ERR_MEMORY_ALLOCATION_FAILED",
						   message, sizeof message - 1, false);
	}
	return buffer;
}
// Where Buffer.indexOf starts, as Node v24.19.0's bidirectionalIndexOf and
// then IndexOfOffset place it for a forward search: the offset is clamped to
// an int32's range, an absent or NaN one is 0, and the rest truncates toward
// zero. A negative one counts from the end, floored at 0. Past where the
// needle can fit, an empty needle starts at the length and any other finds
// nothing, which is false.
static bool index_of_start(double offset, bool has_offset, size_t length,
						   size_t needle_length, size_t *start)
{
	if (!has_offset || isnan(offset)) {
		offset = 0;
	} else if (offset > 2147483647.0) {
		offset = 2147483647.0;
	} else if (offset < -2147483648.0) {
		offset = -2147483648.0;
	}
	double whole = trunc(offset), size = (double)length;
	if (whole < 0) {
		*start = whole + size >= 0 ? (size_t)(whole + size) : 0;
		return needle_length == 0 || *start + needle_length <= length;
	}
	if (whole + (double)needle_length <= size) {
		*start = (size_t)whole;
		return true;
	}
	*start = length;
	return needle_length == 0;
}
// The first match of bytes at or past start, or -1: memchr for the first byte,
// then a compare of the rest. It reads the view's data and length only, so a
// slice searches its own bytes, measured from its own start.
static double index_of_bytes(const adamic_typed_array *buffer, size_t start,
							 const unsigned char *bytes, size_t length)
{
	const unsigned char *data = buffer->data;
	size_t last = buffer->length - length;
	for (size_t at = start; at <= last;) {
		const unsigned char *found = memchr(data + at, bytes[0], last - at + 1);
		if (found == NULL) {
			return -1;
		}
		at = (size_t)(found - data);
		if (memcmp(found + 1, bytes + 1, length - 1) == 0) {
			return (double)at;
		}
		at++;
	}
	return -1;
}
// Buffer.indexOf of a string, searched as its UTF-8 bytes, which are Adamic's
// own: lowering admits only a literal free of lone surrogates (Node v24.19.0
// aborts on one), so its WTF-8 is plain UTF-8, the bytes Node searches.
double adamic_node_buffer_index_of_string(const adamic_typed_array *buffer,
										  const adamic_string *needle,
										  double offset, bool has_offset)
{
	size_t start;
	if (!index_of_start(offset, has_offset, buffer->length, needle->length,
						&start)) {
		return -1;
	}
	if (needle->length == 0) {
		return (double)start;
	}
	return index_of_bytes(buffer, start, (const unsigned char *)needle->bytes,
						  needle->length);
}
// Buffer.indexOf of a number: one byte, Node's val >>> 0 narrowed by memchr to
// its low 8 bits, so 266 is 10, -1 is 255, and NaN and the infinities are 0.
double adamic_node_buffer_index_of_number(const adamic_typed_array *buffer,
										  double needle, double offset,
										  bool has_offset)
{
	size_t start;
	if (!index_of_start(offset, has_offset, buffer->length, 1, &start)) {
		return -1;
	}
	unsigned char byte =
		(unsigned char)((uint32_t)adamic_shift_right_unsigned(needle, 0) & 255);
	return index_of_bytes(buffer, start, &byte, 1);
}
static int unbase64(unsigned c)
{
	if (c >= 'A' && c <= 'Z') {
		return (int)(c - 'A');
	}
	if (c >= 'a' && c <= 'z') {
		return (int)(c - 'a') + 26;
	}
	if (c >= '0' && c <= '9') {
		return (int)(c - '0') + 52;
	}
	if (c == '+' || c == '-') {
		return 62;
	}
	if (c == '/' || c == '_') {
		return 63;
	}
	return -1;
}
static int unhex(unsigned c)
{
	if (c >= '0' && c <= '9') {
		return (int)(c - '0');
	}
	if (c >= 'a' && c <= 'f') {
		return (int)(c - 'a') + 10;
	}
	if (c >= 'A' && c <= 'F') {
		return (int)(c - 'A') + 10;
	}
	return -1;
}
// A typed array's length is fixed, so base64 and hex decode twice: once with
// out NULL to count the bytes, then into a Buffer made at that count. Each
// pass runs the same loop, so the count is exact.
static void put_byte(uint8_t *out, size_t *written, unsigned byte)
{
	if (out != NULL) {
		out[*written] = (uint8_t)(byte & 255);
	}
	*written += 1;
}
static size_t decode_base64(const adamic_string *text, size_t units,
							uint8_t *out)
{
	// Node's Base64DecodeGroupSlow narrows each UTF-16 unit to uint8_t,
	// skips all invalid characters, stops at '=', and emits after sextets
	// 2/3/4.
	size_t written = 0;
	unsigned hi = 0, group = 0;
	for (size_t at = 0; at < units; at++) {
		unsigned c =
			(unsigned)adamic_string_char_code_at(text, (double)at) & 255;
		if (c == '=') {
			break;
		}
		int lo = unbase64(c);
		if (lo < 0) {
			continue;
		}
		if (group == 1) {
			put_byte(out, &written, (hi << 2) | ((unsigned)lo >> 4));
		}
		if (group == 2) {
			put_byte(out, &written, (hi << 4) | ((unsigned)lo >> 2));
		}
		if (group == 3) {
			put_byte(out, &written, (hi << 6) | (unsigned)lo);
		}
		hi = (unsigned)lo;
		group = (group + 1) % 4;
	}
	return written;
}
static size_t decode_hex(const adamic_string *text, size_t units, uint8_t *out)
{
	size_t written = 0;
	for (size_t at = 0; at + 1 < units; at += 2) {
		int hi = unhex(
			(unsigned)adamic_string_char_code_at(text, (double)at) & 255);
		int lo = unhex(
			(unsigned)adamic_string_char_code_at(text, (double)at + 1) & 255);
		if (hi < 0 || lo < 0) {
			break;
		}
		put_byte(out, &written, ((unsigned)hi << 4) | (unsigned)lo);
	}
	return written;
}
adamic_typed_array *adamic_node_buffer_from(const adamic_string *text,
											int encoding)
{
	if (encoding == 0) {
		adamic_typed_array *buffer =
			adamic_typed_array_make(adamic_typed_array_uint8, text->length);
		uint8_t *out = buffer->data;
		for (size_t at = 0; at < text->length; at++) {
			out[at] = (uint8_t)((unsigned)adamic_utf8_at(text, (double)at) & 255);
		}
		return buffer;
	}
	size_t units = adamic_string_units(text);
	if (encoding == 1) {
		// Two bytes a unit. On wasm32 a wrapped size would make a Buffer
		// shorter than the stores below, so the multiply is checked.
		if (units > SIZE_MAX / 2) {
			static const char message[] = "RangeError: Invalid typed array length";
			adamic_panic(message, sizeof message - 1);
		}
		adamic_typed_array *buffer =
			adamic_typed_array_make(adamic_typed_array_uint8, units * 2);
		uint8_t *out = buffer->data;
		for (size_t at = 0; at < units; at++) {
			unsigned unit =
				(unsigned)adamic_string_char_code_at(text, (double)at);
			out[at * 2] = (uint8_t)(unit & 255);
			out[at * 2 + 1] = (uint8_t)((unit >> 8) & 255);
		}
		return buffer;
	}
	size_t (*decode)(const adamic_string *, size_t, uint8_t *) =
		encoding == 2 ? decode_base64 : decode_hex;
	adamic_typed_array *buffer = adamic_typed_array_make(
		adamic_typed_array_uint8, decode(text, units, NULL));
	decode(text, units, buffer->data);
	return buffer;
}
static size_t clamp_offset(double offset, size_t length)
{
	if (!(offset > 0)) {
		return 0;
	}
	if (offset >= (double)length) {
		return length;
	}
	return (size_t)offset;
}
adamic_string *adamic_node_buffer_string(const adamic_typed_array *buffer,
										 double from, double to, int encoding)
{
	size_t start = clamp_offset(from, buffer->length),
		   end = clamp_offset(to, buffer->length);
	if (end <= start) {
		return adamic_retain(&adamic_string_empty);
	}
	size_t length = end - start;
	const uint8_t *bytes = buffer->data;
	if (encoding == 1) {
		adamic_string *text = adamic_string_allocate((length / 2) * 3);
		size_t written = 0;
		for (size_t at = start; at + 1 < end; at += 2) {
			uint32_t point = (unsigned)bytes[at] | ((unsigned)bytes[at + 1] << 8);
			// JavaScript's UTF-16 permits lone surrogates. Store those in
			// WTF-8, joining only a high surrogate immediately followed by a
			// low surrogate.
			if (point >= 0xd800 && point <= 0xdbff && at + 3 < end) {
				unsigned low =
					(unsigned)bytes[at + 2] | ((unsigned)bytes[at + 3] << 8);
				if (low >= 0xdc00 && low <= 0xdfff) {
					point = 0x10000 + ((point - 0xd800) << 10) + low - 0xdc00;
					at += 2;
				}
			}
			written += put_point((unsigned char *)text->bytes + written, point);
		}
		text->length = written;
		return text;
	}
	if (encoding == 3) {
		static const char hex[] = "0123456789abcdef";
		adamic_string *text = adamic_string_allocate(length * 2);
		char *out = (char *)text->bytes;
		for (size_t at = 0; at < length; at++) {
			unsigned byte = bytes[start + at];
			out[at * 2] = hex[byte >> 4];
			out[at * 2 + 1] = hex[byte & 15];
		}
		return text;
	}
	if (encoding == 4) {
		// 'ascii' reads each byte as its low seven bits, as Node does, so 0x80
		// is U+0000 and 0xff U+007F. The text is one-byte UTF-8 throughout, a
		// unit a byte, so units is length + 1, adamic.h's flag for that.
		adamic_string *text = adamic_string_allocate(length);
		char *out = (char *)text->bytes;
		for (size_t at = 0; at < length; at++) {
			out[at] = (char)(bytes[start + at] & 0x7f);
		}
		text->units = length + 1;
		return text;
	}
	if (encoding == 2) {
		static const char alphabet[] =
			"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
		adamic_string *text = adamic_string_allocate(((length + 2) / 3) * 4);
		char *out = (char *)text->bytes;
		size_t written = 0;
		for (size_t at = 0; at < length; at += 3) {
			unsigned a = bytes[start + at];
			unsigned b = at + 1 < length ? bytes[start + at + 1] : 0;
			unsigned c = at + 2 < length ? bytes[start + at + 2] : 0;
			out[written++] = alphabet[a >> 2];
			out[written++] = alphabet[((a & 3) << 4) | (b >> 4)];
			out[written++] =
				at + 1 < length ? alphabet[((b & 15) << 2) | (c >> 6)] : '=';
			out[written++] = at + 2 < length ? alphabet[c & 63] : '=';
		}
		return text;
	}
	return decode_utf8(bytes + start, length);
}
