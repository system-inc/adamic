// node_path.c: POSIX lexical paths. No filesystem lookup, including for NUL
// bytes.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static void *allocate(size_t size) {
	void *result = malloc(size);
	if (result == NULL) {
		static const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	return result;
}
static adamic_string *text(const char *bytes, size_t length) {
	adamic_string *result =
		adamic_allocate(sizeof *result + length, adamic_kind_string);
	result->length = length;
	result->bytes = (const char *)(result + 1);
	result->units = 0;
	result->index = NULL;
	result->owner = NULL;
	result->capacity = length;
	memcpy((char *)result->bytes, bytes, length);
	return result;
}
// Node v24.19.0 lib/path.js normalizeString and posix.normalize (MIT).
// Slash and dot are ASCII; scanning WTF-8 bytes preserves all other code units.
static adamic_string *normalize(const char *bytes, size_t length, bool absolute,
                                bool trailing) {
    char *output = allocate(length + 3);
    size_t used = 0, last_segment_length = 0;
    int64_t last_slash = -1, dots = 0;
    char code = 0;
    for (size_t i = 0; i <= length; i++) {
        if (i < length) code = bytes[i];
        else if (code == '/') break;
        else code = '/';
        if (code == '/') {
            if (last_slash == (int64_t)i - 1 || dots == 1) {
                // Repeated slash or a single dot.
            } else if (dots == 2) {
                if (used < 2 || last_segment_length != 2 ||
                    output[used - 1] != '.' || output[used - 2] != '.') {
                    if (used > 2) {
                        int64_t slash = (int64_t)used - (int64_t)last_segment_length - 1;
                        if (slash == -1) { used = 0; last_segment_length = 0; }
                        else {
                            used = (size_t)slash;
                            int64_t previous = slash - 1;
                            while (previous >= 0 && output[previous] != '/') previous--;
                            last_segment_length = used - 1 - previous;
                        }
                        last_slash = (int64_t)i; dots = 0; continue;
                    } else if (used != 0) {
                        used = 0; last_segment_length = 0;
                        last_slash = (int64_t)i; dots = 0; continue;
                    }
                }
                if (!absolute) {
                    if (used) output[used++] = '/';
                    output[used++] = '.'; output[used++] = '.';
                    last_segment_length = 2;
                }
            } else {
                if (used) output[used++] = '/';
                size_t begin = (size_t)(last_slash + 1);
                last_segment_length = i - begin;
                memcpy(output + used, bytes + begin, last_segment_length);
                used += last_segment_length;
            }
            last_slash = (int64_t)i; dots = 0;
        } else if (code == '.' && dots != -1) dots++;
        else dots = -1;
    }
    if (absolute) {
        memmove(output + 1, output, used); output[0] = '/'; used++;
    } else if (used == 0) output[used++] = '.';
    if (trailing && !(absolute && used == 1)) output[used++] = '/';
    adamic_string *result = text(output, used);
    free(output);
    return result;
}

adamic_string *adamic_node_path_normalize(const adamic_string *path) {
    if (path->length == 0) return text(".", 1);
    return normalize(path->bytes, path->length, path->bytes[0] == '/',
                     path->bytes[path->length - 1] == '/');
}

bool adamic_node_path_isAbsolute(const adamic_string *path) {
    return path->length > 0 && path->bytes[0] == '/';
}

// Node v24.19.0 lib/path.js posix.extname (MIT).
adamic_string *adamic_node_path_extname(const adamic_string *path) {
    int64_t start_dot = -1, start_part = 0, end = -1, pre_dot_state = 0;
    bool matched_slash = true;
    for (int64_t i = (int64_t)adamic_string_length(path) - 1; i >= 0; i--) {
        double code = adamic_string_char_code_at(path, (double)i);
        if (code == '/') {
            if (!matched_slash) { start_part = i + 1; break; }
            continue;
        }
        if (end == -1) { matched_slash = false; end = i + 1; }
        if (code == '.') {
            if (start_dot == -1) start_dot = i;
            else if (pre_dot_state != 1) pre_dot_state = 1;
        } else if (start_dot != -1) pre_dot_state = -1;
    }
    if (start_dot == -1 || end == -1 || pre_dot_state == 0 ||
        (pre_dot_state == 1 && start_dot == end - 1 && start_dot == start_part + 1))
        return text("", 0);
    return adamic_string_slice(path, (double)start_dot, (double)end, true);
}

adamic_string *adamic_node_path_join(size_t count,
									 adamic_string *const paths[]) {
	size_t length = 0;
	for (size_t i = 0; i < count; i++) {
		if (paths[i]->length > 0) {
			length += paths[i]->length + 1;
		}
	}
	if (length == 0) {
		return text(".", 1);
	}
	char *bytes = allocate(length);
	size_t used = 0;
	for (size_t i = 0; i < count; i++) {
		if (paths[i]->length == 0) {
			continue;
		}
		if (used > 0) {
			bytes[used++] = '/';
		}
		memcpy(bytes + used, paths[i]->bytes, paths[i]->length);
		used += paths[i]->length;
	}
	adamic_string *result =
		normalize(bytes, used, bytes[0] == '/', bytes[used - 1] == '/');
	free(bytes);
	return result;
}
adamic_string *adamic_node_path_resolve(size_t count,
										adamic_string *const paths[]) {
	size_t first = 0;
	bool absolute = false;
	for (size_t i = count; i > 0; i--) {
		if (paths[i - 1]->length > 0 && paths[i - 1]->bytes[0] == '/') {
			first = i - 1;
			absolute = true;
			break;
		}
	}
	char *cwd = NULL;
	size_t length = 0;
	if (!absolute) {
		cwd = getcwd(NULL, 0);
		if (cwd == NULL) {
			adamic_node_fs_raise(NULL, errno, "uv_cwd");
			return NULL;
		}
		length = strlen(cwd) + 1;
	}
	for (size_t i = first; i < count; i++) {
		length += paths[i]->length + 1;
	}
	char *bytes = allocate(length + 1);
	size_t used = 0;
	if (cwd != NULL) {
		used = strlen(cwd);
		memcpy(bytes, cwd, used);
		free(cwd);
	}
	for (size_t i = first; i < count; i++) {
		if (paths[i]->length == 0) {
			continue;
		}
		if (used > 0) {
			bytes[used++] = '/';
		}
		memcpy(bytes + used, paths[i]->bytes, paths[i]->length);
		used += paths[i]->length;
	}
	adamic_string *result = normalize(bytes, used, true, false);
	free(bytes);
	return result;
}
adamic_string *adamic_node_path_dirname(const adamic_string *path) {
	if (path->length == 0) {
		return text(".", 1);
	}
	bool root = path->bytes[0] == '/', matched = true;
	size_t end = 0;
	for (size_t i = path->length; i > 1; i--) {
		if (path->bytes[i - 1] == '/') {
			if (!matched) {
				end = i - 1;
				break;
			}
		} else {
			matched = false;
		}
	}
	if (end == 0) {
		return text(root ? "/" : ".", 1);
	}
	if (root && end == 1) {
		return text("//", 2);
	}
	return text(path->bytes, end);
}

// Compare normalized components, not a byte prefix that might end inside a
// name.
adamic_string *adamic_node_path_relative(const adamic_string *from,
										 const adamic_string *to) {
	if (from->length == to->length &&
		memcmp(from->bytes, to->bytes, from->length) == 0) {
		return text("", 0);
	}
	adamic_string *absolute_from = adamic_node_path_resolve(
		1, (adamic_string *const[]){(adamic_string *)from});
	if (absolute_from == NULL) {
		return NULL;
	}
	adamic_string *absolute_to = adamic_node_path_resolve(
		1, (adamic_string *const[]){(adamic_string *)to});
	if (absolute_to == NULL) {
		adamic_release(absolute_from);
		return NULL;
	}
	size_t from_at = 1, to_at = 1;
	while (from_at < absolute_from->length && to_at < absolute_to->length) {
		size_t from_end = from_at, to_end = to_at;
		while (from_end < absolute_from->length &&
			   absolute_from->bytes[from_end] != '/') {
			from_end++;
		}
		while (to_end < absolute_to->length &&
			   absolute_to->bytes[to_end] != '/') {
			to_end++;
		}
		size_t length = from_end - from_at;
		if (length != to_end - to_at ||
			memcmp(absolute_from->bytes + from_at, absolute_to->bytes + to_at,
				   length) != 0) {
			break;
		}
		from_at = from_end + (from_end < absolute_from->length);
		to_at = to_end + (to_end < absolute_to->length);
	}
	size_t parents = 0;
	for (size_t at = from_at; at < absolute_from->length; at++) {
		if (at == from_at || absolute_from->bytes[at - 1] == '/') {
			parents++;
		}
	}
	size_t rest = absolute_to->length - to_at;
	char *bytes = allocate(parents * 3 + rest + 1);
	size_t used = 0;
	for (size_t i = 0; i < parents; i++) {
		if (used > 0) {
			bytes[used++] = '/';
		}
		bytes[used++] = '.';
		bytes[used++] = '.';
	}
	if (rest > 0) {
		if (used > 0) {
			bytes[used++] = '/';
		}
		memcpy(bytes + used, absolute_to->bytes + to_at, rest);
		used += rest;
	}
	adamic_string *result = text(bytes, used);
	free(bytes);
	adamic_release(absolute_from);
	adamic_release(absolute_to);
	return result;
}

// Ported from Node v24.19.0 lib/path.js, posix.basename (MIT; see notices).
// Scan UTF-16 units: a suffix may match half of a surrogate pair.
adamic_string *adamic_node_path_basename(const adamic_string *path,
                                       const adamic_string *suffix) {
    int64_t length = (int64_t)adamic_string_length(path);
    int64_t start = 0, end = -1;
    bool matched_slash = true;
    int64_t suffix_length = suffix == NULL ? 0 : (int64_t)adamic_string_length(suffix);
    if (suffix_length > 0 && suffix_length <= length) {
        if (adamic_string_equal(suffix, path)) return text("", 0);
        int64_t extension_at = suffix_length - 1, first_non_slash_end = -1;
        for (int64_t i = length - 1; i >= 0; i--) {
            double code = adamic_string_char_code_at(path, (double)i);
            if (code == '/') {
                if (!matched_slash) { start = i + 1; break; }
            } else {
                if (first_non_slash_end == -1) {
                    matched_slash = false;
                    first_non_slash_end = i + 1;
                }
                if (extension_at >= 0) {
                    if (code == adamic_string_char_code_at(suffix, (double)extension_at)) {
                        if (--extension_at == -1) end = i;
                    } else {
                        extension_at = -1;
                        end = first_non_slash_end;
                    }
                }
            }
        }
        if (start == end) end = first_non_slash_end;
        else if (end == -1) end = length;
        return adamic_string_slice(path, (double)start, (double)end, true);
    }
    for (int64_t i = length - 1; i >= 0; i--) {
        if (adamic_string_char_code_at(path, (double)i) == '/') {
            if (!matched_slash) { start = i + 1; break; }
        } else if (end == -1) {
            matched_slash = false;
            end = i + 1;
        }
    }
    if (end == -1) return text("", 0);
    return adamic_string_slice(path, (double)start, (double)end, true);
}
