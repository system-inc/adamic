// Private string builder implementation, included only by string.c.

// A growing buffer for building strings, WTF-8 included.
typedef struct builder {
	char *bytes;
	size_t length;
	size_t capacity;
} builder;

static void builder_add(builder *build, const char *bytes, size_t size) {
	if (build->length + size > build->capacity) {
		size_t capacity = build->capacity == 0 ? 32 : build->capacity;
		while (capacity < build->length + size) {
			capacity *= 2;
		}
		char *grown = realloc(build->bytes, capacity);
		if (grown == NULL) {
			static const char message[] = "out of memory";
			adamic_panic(message, sizeof message - 1);
		}
		build->bytes = grown;
		build->capacity = capacity;
	}
	if (size > 0) {
		memcpy(build->bytes + build->length, bytes, size);
	}
	build->length += size;
}

// builder_unit appends one UTF-16 code unit: a BMP character, or a lone surrogate as WTF-8's three
// bytes (ED A0 80 to ED BF BF), which no valid UTF-8 contains, so it can never be mistaken.
static void builder_unit(builder *build, unsigned unit) {
	char bytes[3];
	if (unit < 0x80) {
		bytes[0] = (char)unit;
		builder_add(build, bytes, 1);
	} else if (unit < 0x800) {
		bytes[0] = (char)(0xc0 | (unit >> 6));
		bytes[1] = (char)(0x80 | (unit & 0x3f));
		builder_add(build, bytes, 2);
	} else {
		bytes[0] = (char)(0xe0 | (unit >> 12));
		bytes[1] = (char)(0x80 | ((unit >> 6) & 0x3f));
		bytes[2] = (char)(0x80 | (unit & 0x3f));
		builder_add(build, bytes, 3);
	}
}

static adamic_string *builder_finish_units(builder *build, size_t units) {
	adamic_string piece = {{0, adamic_kind_string, 0}, build->length, build->bytes, units, NULL, NULL, 0};
	adamic_string *string = adamic_string_concat(1, (adamic_string *const[]){&piece});
	free(build->bytes);
	return string;
}


static adamic_string *builder_finish(builder *build) {
	return builder_finish_units(build, 0);
}
