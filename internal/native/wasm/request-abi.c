// Temporary export adapter for the executable C backend. The wasm compiler will emit this ABI.
#define main adamic_fixture_main
#include "program.c"
#undef main
#include "count.h"

// Input is borrowed UTF-8 for this call. The returned string is owned by the host.
__attribute__((export_name("adamic_request")))
adamic_string *adamic_request(const unsigned char *bytes, size_t length) {
	adamic_string *request = adamic_decode_utf8(bytes, length);
	adamic_string *response = ADAMIC_HANDLER(request);
	adamic_release(request);
	return response;
}

__attribute__((export_name("adamic_response_bytes")))
const char *adamic_response_bytes(const adamic_string *response) {
	return response->bytes;
}

__attribute__((export_name("adamic_response_length")))
size_t adamic_response_length(const adamic_string *response) {
	return response->length;
}

// Counted-build probes distinguish live allocations from reusable linear-memory capacity.
__attribute__((export_name("adamic_live")))
size_t adamic_live(void) { return adamic_counted.live; }

__attribute__((export_name("adamic_regions")))
size_t adamic_regions(void) { return adamic_counted.regions; }
