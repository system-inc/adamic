// SHA-256: FIPS 180-4 sections 4.1.2, 4.2.2, 5.1.1, 5.2.1, 5.3.3 and 6.2.
// The state and pending bytes live in ordinary counted numeric arrays, so the
// existing heap destruction paths release everything without a new heap kind.
#include "adamic.h"
#include <string.h>

static const uint32_t constants[64] = {
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1,
	0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
	0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174, 0xe49b69c1, 0xefbe4786,
	0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147,
	0x06ca6351, 0x14292967, 0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
	0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85, 0xa2bfe8a1, 0xa81a664b,
	0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a,
	0x5b9cca4f, 0x682e6ff3, 0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
	0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2};
static uint32_t rotate(uint32_t word, unsigned shift)
{
	return (word >> shift) | (word << (32 - shift));
}
static void compress(uint32_t state[8], const unsigned char bytes[64])
{
	uint32_t words[64];
	for (size_t i = 0; i < 16; i++) {
		words[i] = ((uint32_t)bytes[i * 4] << 24) |
				   ((uint32_t)bytes[i * 4 + 1] << 16) |
				   ((uint32_t)bytes[i * 4 + 2] << 8) | bytes[i * 4 + 3];
	}
	for (size_t i = 16; i < 64; i++) {
		uint32_t a = words[i - 15], b = words[i - 2];
		words[i] = words[i - 16] + (rotate(a, 7) ^ rotate(a, 18) ^ (a >> 3)) +
				   words[i - 7] + (rotate(b, 17) ^ rotate(b, 19) ^ (b >> 10));
	}
	uint32_t a = state[0], b = state[1], c = state[2], d = state[3],
			 e = state[4], f = state[5], g = state[6], h = state[7];
	for (size_t i = 0; i < 64; i++) {
		uint32_t first = h + (rotate(e, 6) ^ rotate(e, 11) ^ rotate(e, 25)) +
						 ((e & f) ^ (~e & g)) + constants[i] + words[i];
		uint32_t second = (rotate(a, 2) ^ rotate(a, 13) ^ rotate(a, 22)) +
						  ((a & b) ^ (a & c) ^ (b & c));
		h = g;
		g = f;
		f = e;
		e = d + first;
		d = c;
		c = b;
		b = a;
		a = first + second;
	}
	state[0] += a;
	state[1] += b;
	state[2] += c;
	state[3] += d;
	state[4] += e;
	state[5] += f;
	state[6] += g;
	state[7] += h;
}
static const char *const names[] = {"bytes", "finalized"};
static const bool references[] = {true, false};
static const adamic_shape shape = {2, names, references, NULL, NULL};
adamic_object *adamic_node_hash_new(void)
{
	adamic_object *hash = adamic_object_new(&shape);
	hash->slots[0].reference = adamic_array_new(0, false);
	return hash;
}
static void check_finalized(const adamic_object *hash)
{
	if (hash->slots[1].number != 0) {
		static const char message[] =
			"Error [ERR_CRYPTO_HASH_FINALIZED]: Digest already called";
		adamic_panic(message, sizeof message - 1);
	}
}
adamic_object *adamic_node_hash_update(adamic_object *hash,
									   const adamic_string *text, int encoding)
{
	check_finalized(hash);
	adamic_array *input = adamic_node_buffer_from(text, encoding);
	adamic_array *bytes = hash->slots[0].reference;
	for (size_t i = 0; i < input->length; i++) {
		adamic_array_push(bytes, input->elements[i]);
	}
	adamic_release(input);
	return adamic_retain(hash);
}
adamic_string *adamic_node_hash_digest(adamic_object *hash)
{
	check_finalized(hash);
	hash->slots[1].number = 1;
	const adamic_array *bytes = hash->slots[0].reference;
	uint32_t state[8] = {0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
						 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19};
	unsigned char block[64];
	size_t at = 0;
	while (bytes->length - at >= 64) {
		for (size_t i = 0; i < 64; i++) {
			block[i] = (unsigned char)bytes->elements[at + i].number;
		}
		compress(state, block);
		at += 64;
	}
	size_t remaining = bytes->length - at;
	memset(block, 0, sizeof block);
	for (size_t i = 0; i < remaining; i++) {
		block[i] = (unsigned char)bytes->elements[at + i].number;
	}
	block[remaining] = 0x80;
	if (remaining >= 56) {
		compress(state, block);
		memset(block, 0, sizeof block);
	}
	uint64_t bits = (uint64_t)bytes->length * 8;
	for (size_t i = 0; i < 8; i++) {
		block[63 - i] = (unsigned char)(bits >> (i * 8));
	}
	compress(state, block);
	static const char hex[] = "0123456789abcdef";
	adamic_string *digest = adamic_string_allocate(64);
	for (size_t i = 0; i < 32; i++) {
		unsigned byte = (state[i / 4] >> ((3 - i % 4) * 8)) & 255;
		((char *)digest->bytes)[i * 2] = hex[byte >> 4];
		((char *)digest->bytes)[i * 2 + 1] = hex[byte & 15];
	}
	return digest;
}
