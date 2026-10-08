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
static void push_byte(adamic_array *array, unsigned byte)
{
	adamic_array_push(array, (adamic_value){.number = (double)(byte & 255)});
}
static unsigned byte_value(double value)
{
	if (!isfinite(value) || value == 0) {
		return 0;
	}
	double byte = fmod(trunc(value), 256);
	if (byte < 0) {
		byte += 256;
	}
	return (unsigned)byte;
}
adamic_array *adamic_node_buffer_copy(const adamic_array *source)
{
	adamic_array *buffer = adamic_array_new(source->length, false);
	for (size_t at = 0; at < source->length; at++) {
		push_byte(buffer, byte_value(source->elements[at].number));
	}
	return buffer;
}
double adamic_node_buffer_set(adamic_array *buffer, double index, double value)
{
	adamic_value *slot = adamic_array_at(buffer, index);
	// Uint8Array writes outside the bounds have no effect, unlike Array writes.
	if (slot != NULL) {
		slot->number = (double)byte_value(value);
	}
	return value;
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
adamic_array *adamic_node_buffer_from(const adamic_string *text, int encoding)
{
	size_t units = adamic_string_units(text);
	adamic_array *buffer =
		adamic_array_new(encoding == 1 ? units * 2 : text->length, false);
	if (encoding == 0) {
		for (size_t at = 0; at < text->length; at++) {
			push_byte(buffer, (unsigned)adamic_utf8_at(text, (double)at));
		}
	} else if (encoding == 1) {
		for (size_t at = 0; at < units; at++) {
			unsigned unit =
				(unsigned)adamic_string_char_code_at(text, (double)at);
			push_byte(buffer, unit);
			push_byte(buffer, unit >> 8);
		}
	} else if (encoding == 4) {
		for (size_t at = 0; at < units; at++) {
			push_byte(buffer, (unsigned)adamic_string_char_code_at(text, (double)at));
		}
	} else if (encoding == 2) {
		// Node's Base64DecodeGroupSlow narrows each UTF-16 unit to uint8_t,
		// skips all invalid characters, stops at '=', and emits after sextets
		// 2/3/4.
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
				push_byte(buffer, (hi << 2) | ((unsigned)lo >> 4));
			}
			if (group == 2) {
				push_byte(buffer, (hi << 4) | ((unsigned)lo >> 2));
			}
			if (group == 3) {
				push_byte(buffer, (hi << 6) | (unsigned)lo);
			}
			hi = (unsigned)lo;
			group = (group + 1) % 4;
		}
	} else {
		for (size_t at = 0; at + 1 < units; at += 2) {
			int hi = unhex(
				(unsigned)adamic_string_char_code_at(text, (double)at) & 255);
			int lo = unhex(
				(unsigned)adamic_string_char_code_at(text, (double)at + 1) &
				255);
			if (hi < 0 || lo < 0) {
				break;
			}
			push_byte(buffer, ((unsigned)hi << 4) | (unsigned)lo);
		}
	}
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
adamic_string *adamic_node_buffer_string(const adamic_array *buffer,
										 double from, double to, int encoding)
{
	size_t start = clamp_offset(from, buffer->length),
		   end = clamp_offset(to, buffer->length);
	if (end <= start) {
		return adamic_retain(&adamic_string_empty);
	}
	size_t length = end - start;
	if (encoding == 1) {
		adamic_string *text = adamic_string_allocate((length / 2) * 3);
		size_t written = 0;
		for (size_t at = start; at + 1 < end; at += 2) {
			uint32_t point = (unsigned)buffer->elements[at].number |
							 ((unsigned)buffer->elements[at + 1].number << 8);
			// JavaScript's UTF-16 permits lone surrogates. Store those in
			// WTF-8, joining only a high surrogate immediately followed by a
			// low surrogate.
			if (point >= 0xd800 && point <= 0xdbff && at + 3 < end) {
				unsigned low = (unsigned)buffer->elements[at + 2].number |
							   ((unsigned)buffer->elements[at + 3].number << 8);
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
	if (encoding == 4) {
		adamic_string *text = adamic_string_allocate(length * 2);
		size_t written = 0;
		for (size_t at = start; at < end; at++) {
			written += put_point((unsigned char *)text->bytes + written,
				(unsigned)buffer->elements[at].number);
		}
		text->length = written;
		return text;
	}
	if (encoding == 3) {
		static const char hex[] = "0123456789abcdef";
		adamic_string *text = adamic_string_allocate(length * 2);
		char *out = (char *)text->bytes;
		for (size_t at = 0; at < length; at++) {
			unsigned byte = (unsigned)buffer->elements[start + at].number;
			out[at * 2] = hex[byte >> 4];
			out[at * 2 + 1] = hex[byte & 15];
		}
		return text;
	}
	if (encoding == 2) {
		static const char alphabet[] =
			"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
		adamic_string *text = adamic_string_allocate(((length + 2) / 3) * 4);
		char *out = (char *)text->bytes;
		size_t written = 0;
		for (size_t at = 0; at < length; at += 3) {
			unsigned a = (unsigned)buffer->elements[start + at].number;
			unsigned b = at + 1 < length
							 ? (unsigned)buffer->elements[start + at + 1].number
							 : 0;
			unsigned c = at + 2 < length
							 ? (unsigned)buffer->elements[start + at + 2].number
							 : 0;
			out[written++] = alphabet[a >> 2];
			out[written++] = alphabet[((a & 3) << 4) | (b >> 4)];
			out[written++] =
				at + 1 < length ? alphabet[((b & 15) << 2) | (c >> 6)] : '=';
			out[written++] = at + 2 < length ? alphabet[c & 63] : '=';
		}
		return text;
	}
	unsigned char *bytes = malloc(length);
	if (bytes == NULL) {
		adamic_panic("out of memory", 13);
	}
	for (size_t at = 0; at < length; at++) {
		bytes[at] = (unsigned char)buffer->elements[start + at].number;
	}
	adamic_string *text = decode_utf8(bytes, length);
	free(bytes);
	return text;
}

// Node Buffer slice/subarray offsets use ToInteger and permit negative offsets.
static size_t view_offset(double offset, size_t length)
{
	if (isnan(offset)) {
		return 0;
	}
	offset = trunc(offset);
	if (offset < 0) {
		offset += (double)length;
	}
	if (offset <= 0) {
		return 0;
	}
	if (offset >= (double)length) {
		return length;
	}
	return (size_t)offset;
}
adamic_array *adamic_node_buffer_view(adamic_array *buffer, double from, double to)
{
	size_t start = view_offset(from, buffer->length),
		   end = view_offset(to, buffer->length);
	if (end < start) {
		end = start;
	}
	adamic_array *view = adamic_array_new(0, false);
	view->length = end - start;
	view->capacity = view->length;
	view->elements = buffer->elements == NULL ? NULL : buffer->elements + start;
	view->owner = adamic_retain(buffer->owner != NULL ? buffer->owner : buffer);
	return view;
}
