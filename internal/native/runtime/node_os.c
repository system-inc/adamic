// POSIX home lookup ported from Node v24.19.0's GetHomeDirectory and libuv's
// uv_os_homedir/uv__getpwuid_r. See THIRD_PARTY_NOTICES.md.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include <errno.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#ifndef ADAMIC_TARGET_WASI
#include <pwd.h>
#include <unistd.h>

static const char *const home_error_names[] = {"name", "message", "code", "info", "errno", "syscall"};
static const bool home_error_refs[] = {true, true, true, true, false, true};
static const adamic_shape home_error_shape = {6, home_error_names, home_error_refs, NULL};
static adamic_class home_error_class = {NULL, 3, 6, NULL, 0, &home_error_shape, NULL, 0, false, 0, NULL};

static adamic_string *os_text(const char *value) {
    return adamic_decode_utf8((const unsigned char *)value, strlen(value));
}

static void home_error(int error) {
    const char *code, *reason;
    switch (error) {
    case ENOENT: code="ENOENT"; reason="no such file or directory"; break;
    case ENOBUFS: code="ENOBUFS"; reason="no buffer space available"; break;
    case ENOMEM: code="ENOMEM"; reason="not enough memory"; break;
    case EACCES: code="EACCES"; reason="permission denied"; break;
    case EPERM: code="EPERM"; reason="operation not permitted"; break;
    case EIO: code="EIO"; reason="i/o error"; break;
    case EMFILE: code="EMFILE"; reason="too many open files"; break;
    case ENFILE: code="ENFILE"; reason="file table overflow"; break;
    case EINVAL: code="EINVAL"; reason="invalid argument"; break;
    case ESRCH: code="ESRCH"; reason="no such process"; break;
    default: code="UNKNOWN"; reason="unknown error"; break;
    }
    static const char *const info_names[] = {"errno", "code", "message", "syscall"};
    static const bool info_refs[] = {false, true, true, true};
    static const adamic_shape info_shape = {4, info_names, info_refs, NULL};
    static const int info_types[] = {1, 3, 3, 3};
    static adamic_shape_types info_metadata = {&info_shape, info_types, NULL};
    static const int error_types[] = {3, 3, 3, 5, 1, 3};
    static adamic_shape_types error_metadata = {&home_error_shape, error_types, NULL};
    adamic_register_shape_types(&info_metadata);
    adamic_register_shape_types(&error_metadata);
    adamic_object *info = adamic_object_new(&info_shape);
    info->slots[0].number = -(double)error;
    info->slots[1].reference = os_text(code);
    info->slots[2].reference = os_text(reason);
    info->slots[3].reference = os_text("uv_os_homedir");
    char message[256];
    snprintf(message, sizeof message, "A system error occurred: uv_os_homedir returned %s (%s)", code, reason);
    adamic_object *thrown = adamic_object_new(&home_error_shape);
    thrown->slots[0].reference = os_text("SystemError");
    thrown->slots[1].reference = os_text(message);
    thrown->slots[2].reference = os_text("ERR_SYSTEM_ERROR");
    thrown->slots[3].reference = info;
    thrown->slots[4].number = info->slots[0].number;
    thrown->slots[5].reference = adamic_retain(info->slots[3].reference);
    adamic_error_tag(thrown);
    if (home_error_class.base == NULL) { home_error_class.base = thrown->class; }
    thrown->class = &home_error_class;
    adamic_thrown = thrown;
}

static adamic_string *home_value(const char *value) {
    // Node's C++ binding supplies a fixed PATH_MAX buffer, not a growing one.
    if (strlen(value) >= PATH_MAX) { home_error(ENOBUFS); return NULL; }
    return os_text(value);
}
#endif

adamic_string *adamic_node_homedir(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_panic("wasm32-wasi: os.homedir requires an effective-user account database", sizeof "wasm32-wasi: os.homedir requires an effective-user account database" - 1);
#else
    const char *home = getenv("HOME");
    // An empty HOME is set, and must return the empty string, as libuv does.
    if (home != NULL) { return home_value(home); }
    size_t capacity = 2000;
    for (;;) {
        char *buffer = malloc(capacity);
        if (buffer == NULL) { home_error(ENOMEM); return NULL; }
        struct passwd entry, *found = NULL;
        int error;
        do { error = getpwuid_r(geteuid(), &entry, buffer, capacity, &found); } while (error == EINTR);
        if (error == ERANGE) {
            free(buffer);
            if (capacity > SIZE_MAX / 2) { home_error(ENOMEM); return NULL; }
            capacity *= 2;
            continue;
        }
        if (error != 0 || found == NULL) {
            free(buffer); home_error(error == 0 ? ENOENT : error); return NULL;
        }
        adamic_string *result = home_value(entry.pw_dir);
        free(buffer);
        return result;
    }
#endif
}

bool adamic_node_os_system_error(const adamic_object *error) {
#ifdef ADAMIC_TARGET_WASI
    (void)error;
    return false;
#else
    return error->class == &home_error_class;
#endif
}

adamic_string *adamic_node_os_platform(void) { return adamic_node_platform(); }
