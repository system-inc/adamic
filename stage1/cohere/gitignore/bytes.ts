// What the Go gets for free and Adamic 0.1 does not have: a string's bytes. A Go string is a sequence of
// bytes, and the gitignore Go indexes it byte by byte, as git does. An Adamic string is a sequence of
// UTF-16 code units, and 0.1 has no byte array or encoder (no Uint8Array, no TextEncoder), so the port
// spells a string's UTF-8 as a byte string: one character per byte, each in 0 to 255. Two strings are
// equal exactly when their byte strings are, and one ends with another exactly when theirs does, so the
// glob compares byte strings where the Go compares strings.

import { panic } from 'adamic';

// utf8Bytes is text's UTF-8 as a byte string. ASCII text is its own byte string and is returned as it
// is. A lone surrogate, which UTF-8 cannot hold, is spelled as WTF-8 does, three bytes.
export function utf8Bytes(text: string): string {
	let ascii = true;
	for (let index = 0; index < text.length; index++) {
		if (text.charCodeAt(index) >= 0x80) {
			ascii = false;
			break;
		}
	}
	if (ascii) {
		return text;
	}
	let bytes = '';
	for (const character of text) {
		const codePoint = character.codePointAt(0) ?? panic('bytes: a character with no code point');
		if (codePoint < 0x80) {
			bytes += character;
		} else if (codePoint < 0x800) {
			bytes += String.fromCharCode(0xc0 | (codePoint >> 6), 0x80 | (codePoint & 0x3f));
		} else if (codePoint < 0x10000) {
			bytes += String.fromCharCode(0xe0 | (codePoint >> 12), 0x80 | ((codePoint >> 6) & 0x3f), 0x80 | (codePoint & 0x3f));
		} else {
			bytes += String.fromCharCode(
				0xf0 | (codePoint >> 18),
				0x80 | ((codePoint >> 12) & 0x3f),
				0x80 | ((codePoint >> 6) & 0x3f),
				0x80 | (codePoint & 0x3f),
			);
		}
	}
	return bytes;
}

// byteAt is the byte at index of a byte string, as a one-character string: the Go's text[index], which
// panics past the end as this does.
export function byteAt(bytes: string, index: number): string {
	return bytes[index] ?? panic(`bytes: index ${index} is past the end of ${bytes.length} bytes`);
}
