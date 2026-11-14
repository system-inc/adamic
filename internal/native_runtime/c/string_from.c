// text.indexOf(search, position) and text.includes(search, position): the search begins at the
// position, a UTF-16 index clamped to the text, as ECMAScript's ToIntegerOrInfinity and clamping say.
//
// The search runs in place from the position (adamic_string_index_of_at), with nothing copied, so a
// loop that searches a long text from many positions pays only for what it reads.

#include "adamic.h"

#include <math.h>

double adamic_string_index_of_from(const adamic_string *string, const adamic_string *search, double position) {
	double length = adamic_string_length(string);
	double start = isnan(position) ? 0 : trunc(position);
	start = start < 0 ? 0 : start > length ? length : start;
	return adamic_string_index_of_at(string, search, (size_t)start);
}
