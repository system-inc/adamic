// text.indexOf(search, position) and text.includes(search, position): the search begins at the
// position, a UTF-16 index clamped to the text, as ECMAScript's ToIntegerOrInfinity and clamping say.
//
// The search runs over the rest of the text from the position, made by slice, so a position inside a
// surrogate pair and a search for half of one behave as adamic_string_index_of has them behave from
// the start. That copies the rest once per call: a loop that searches a long text from many positions
// pays it each time.

#include "adamic.h"

#include <math.h>

double adamic_string_index_of_from(const adamic_string *string, const adamic_string *search, double position) {
	double length = adamic_string_length(string);
	double start = isnan(position) ? 0 : trunc(position);
	start = start < 0 ? 0 : start > length ? length : start;
	if (start == 0) {
		return adamic_string_index_of(string, search);
	}
	adamic_string *rest = adamic_string_slice(string, start, length, true);
	double found = adamic_string_index_of(rest, search);
	adamic_release(rest);
	return found < 0 ? -1 : found + start;
}
