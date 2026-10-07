// Private string repeat implementation, included only by string.c.

// repeat_unchecked is count copies of string, count already a whole number in range.
static adamic_string *repeat_unchecked(const adamic_string *string, double count) {
	builder build = {NULL, 0, 0};
	for (double index = 0; index < count; index++) {
		builder_add(&build, string->bytes, string->length);
	}
	return builder_finish_units(&build, (size_t)count * adamic_string_units(string) + 1);
}

adamic_string *adamic_string_repeat(const adamic_string *string, double count) {
	count = isnan(count) ? 0 : trunc(count);
	if (count < 0 || isinf(count)) {
		char message[96];
		char number[ADAMIC_NUMBER_FORMAT_MAX];
		size_t length = adamic_number_format(count, number);
		int written = snprintf(message, sizeof message, "RangeError: Invalid count value: %.*s", (int)length, number);
		adamic_panic(message, (size_t)written);
	}
	// Checked before any of it is built, as V8 does: an empty string repeats to itself however many
	// times, and anything longer than V8's longest string is refused.
	if (string->length == 0 || count == 0) {
		return adamic_retain((adamic_string *)&adamic_string_empty);
	}
	adamic_string_check_length(adamic_string_length(string) * count);
	return repeat_unchecked(string, count);
}

adamic_string *adamic_string_pad(const adamic_string *string, double target, const adamic_string *fill, bool at_start) {
	double length = adamic_string_length(string);
	target = isnan(target) ? 0 : trunc(target);
	if (target <= length || fill->length == 0) {
		return adamic_retain((adamic_string *)string);
	}
	adamic_string_check_length(target);
	// The fill, repeated and cut to exactly the missing number of UTF-16 units: whole fills, then the
	// start of one more, so nothing on the way is longer than the result.
	double missing = target - length;
	double fill_length = adamic_string_length(fill);
	double whole = floor(missing / fill_length);
	adamic_string *repeated = repeat_unchecked(fill, whole);
	adamic_string *rest = adamic_string_slice(fill, 0, missing - whole * fill_length, true);
	adamic_string *parts[3] = {repeated, rest, (adamic_string *)string};
	if (!at_start) {
		parts[0] = (adamic_string *)string;
		parts[1] = repeated;
		parts[2] = rest;
	}
	adamic_string *padded = adamic_string_concat(3, parts);
	adamic_release(repeated);
	adamic_release(rest);
	return padded;
}

