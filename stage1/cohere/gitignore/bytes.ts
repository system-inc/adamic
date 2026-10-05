// What the Go gets for free and Adamic 0.1 does not have: a string's bytes. A Go string is a sequence of
// bytes, and the gitignore Go indexes it byte by byte, as git does. An Adamic string is a sequence of
// UTF-16 code units, and 0.1 has no byte array or encoder (no Uint8Array, no TextEncoder), so the port
// spells a string's UTF-8 as a byte string: one character per byte, each in 0 to 255. Two strings are
// equal exactly when their byte strings are, and one ends with another exactly when theirs does, so the
// glob compares byte strings where the Go compares strings.

import { panic } from 'adamic';

// sixBits is the group of six bits of codePoint that UTF-8 puts in a byte, counting groups from the low
// end: (codePoint >> (6 * group)) & 0x3f. Stage 0 does not lower the bitwise operators yet (gap 2 in
// GAPS.md), so it is arithmetic.
function sixBits(codePoint: number, group: number): number {
	return Math.floor(codePoint / 64 ** group) % 64;
}

const byteCharacters =
	'\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x20\x21\x22\x23\x24\x25\x26\x27\x28\x29\x2a\x2b\x2c\x2d\x2e\x2f\x30\x31\x32\x33\x34\x35\x36\x37\x38\x39\x3a\x3b\x3c\x3d\x3e\x3f\x40\x41\x42\x43\x44\x45\x46\x47\x48\x49\x4a\x4b\x4c\x4d\x4e\x4f\x50\x51\x52\x53\x54\x55\x56\x57\x58\x59\x5a\x5b\x5c\x5d\x5e\x5f\x60\x61\x62\x63\x64\x65\x66\x67\x68\x69\x6a\x6b\x6c\x6d\x6e\x6f\x70\x71\x72\x73\x74\x75\x76\x77\x78\x79\x7a\x7b\x7c\x7d\x7e\x7f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff';

// byteCharacter is String.fromCharCode(value) for a byte. Stage 0 does not lower String.fromCharCode yet
// (gap 1 in GAPS.md), so the character is read from a table of all 256.
function byteCharacter(value: number): string {
	return byteCharacters[value] ?? panic(`bytes: ${value} is not a byte`);
}

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
			bytes += byteCharacter(0xc0 + sixBits(codePoint, 1)) + byteCharacter(0x80 + sixBits(codePoint, 0));
		} else if (codePoint < 0x10000) {
			bytes += byteCharacter(0xe0 + sixBits(codePoint, 2)) + byteCharacter(0x80 + sixBits(codePoint, 1)) + byteCharacter(0x80 + sixBits(codePoint, 0));
		} else {
			bytes +=
				byteCharacter(0xf0 + sixBits(codePoint, 3)) +
				byteCharacter(0x80 + sixBits(codePoint, 2)) +
				byteCharacter(0x80 + sixBits(codePoint, 1)) +
				byteCharacter(0x80 + sixBits(codePoint, 0));
		}
	}
	return bytes;
}

// byteAt is the byte at index of a byte string, as a one-character string: the Go's text[index], which
// panics past the end as this does.
export function byteAt(bytes: string, index: number): string {
	return bytes[index] ?? panic(`bytes: index ${index} is past the end of ${bytes.length} bytes`);
}
