// node_process.c: synchronous compiler host observations and performance entries.
#define _POSIX_C_SOURCE 200809L
#ifdef ADAMIC_TARGET_WASI
#define _XOPEN_SOURCE 700
#endif
// macOS hides malloc/malloc.h's zone types once _POSIX_C_SOURCE is set, unless Darwin's
// own extensions are asked for too. glibc ignores this macro.
#define _DARWIN_C_SOURCE
#include "adamic.h"
#include "host_runtime.h"
#include "node_cwd_wasi.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#ifndef ADAMIC_TARGET_WASI
#include <sys/ioctl.h>
#endif
#include <sys/stat.h>
#include <fcntl.h>
#include <time.h>
#include <unistd.h>
static adamic_array *arguments, *execution_arguments;
static adamic_string *current_directory;
static adamic_object *performance_object;

typedef struct mark_record {
    struct mark_record *next;
    size_t length;
    double time;
    char bytes[];
} mark_record;
static mark_record *marks;

static void clear_marks(const adamic_string *name) {
    mark_record **link = &marks;
    while (*link != NULL) {
        mark_record *record = *link;
        if (name == NULL || (record->length == name->length && memcmp(record->bytes, name->bytes, name->length) == 0)) {
            *link = record->next;
            free(record);
        } else { link = &record->next; }
    }
}

static void finish_host(void) {
    clear_marks(NULL);
    if (performance_object != NULL) { adamic_release(performance_object); performance_object = NULL; }
    if (arguments != NULL) { adamic_release(arguments); }
    if (execution_arguments != NULL) { adamic_release(execution_arguments); }
    if (current_directory != NULL) { adamic_release(current_directory); }
    current_directory = NULL;
    arguments = NULL;
    execution_arguments = NULL;
}

// Runtime owns startup observations; this hook only releases library caches and entries.
void adamic_node_process_start(int count, char **values) {
    (void)count;
    (void)values;
    atexit(finish_host);
}

double adamic_node_performance_now(void) { return adamic_host_performance_now(); }
double adamic_node_time_origin(void) { return adamic_host_time_origin(); }
adamic_string *adamic_node_eol(void) { return adamic_host_eol(); }

bool adamic_node_next_tick_feature(void) { return true; }

double adamic_node_pid(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_panic("wasm32-wasi: process.pid requires process identifiers", sizeof "wasm32-wasi: process.pid requires process identifiers" - 1);
#else
    return (double)getpid();
#endif
}

adamic_string *adamic_node_platform(void) {
#if defined(ADAMIC_TARGET_WASI)
    adamic_panic("wasm32-wasi: process.platform requires a Node host platform", sizeof "wasm32-wasi: process.platform requires a Node host platform" - 1);
#elif defined(__APPLE__)
    static adamic_string platform = ADAMIC_STRING("darwin");
#elif defined(__linux__)
    static adamic_string platform = ADAMIC_STRING("linux");
#else
#error Unsupported Node host platform
#endif
#ifndef ADAMIC_TARGET_WASI
    return &platform;
#endif
}

// Code-bearing host errors use the existing pending-exception word and cleanup paths.
static void host_directory_error(int error, const char *operation, const adamic_string *from, const adamic_string *to) {
    const char *code, *description;
    switch (error) {
        case ENOENT: code = "ENOENT"; description = "no such file or directory"; break;
        case ENOTDIR: code = "ENOTDIR"; description = "not a directory"; break;
        case EACCES: code = "EACCES"; description = "permission denied"; break;
        case ELOOP: code = "ELOOP"; description = "too many symbolic links encountered"; break;
        case ENAMETOOLONG: code = "ENAMETOOLONG"; description = "name too long"; break;
        case ENOMEM: code = "ENOMEM"; description = "not enough memory"; break;
        case EIO: code = "EIO"; description = "i/o error"; break;
        default: { static const char message[] = "unsupported host directory errno"; adamic_panic(message, sizeof message - 1); }
    }
    if (error == ENOENT && strcmp(operation, "uv_cwd") == 0) {
        description = "process.cwd failed with error no such file or directory, the current working directory was likely removed without changing the working directory";
    }
    static const char *const directory_error_names[] = {"name", "message", "code"};
    static const bool references[] = {true, true, true};
    static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference};
    static const adamic_shape shape = {3, directory_error_names, references, NULL, kinds};
    static adamic_string error_name = ADAMIC_STRING("Error");
    adamic_string *prefix = adamic_decode_utf8((const unsigned char *)code, strlen(code));
    adamic_string *reason = adamic_decode_utf8((const unsigned char *)description, strlen(description));
    adamic_string *syscall = adamic_decode_utf8((const unsigned char *)operation, strlen(operation));
    adamic_string colon = ADAMIC_STRING(": "), comma = ADAMIC_STRING(", ");
    adamic_string quote = ADAMIC_STRING("'"), arrow = ADAMIC_STRING("' -> '");
    adamic_string *message;
    if (to == NULL) {
        message = adamic_string_concat(5, (adamic_string *const[]){prefix, &colon, reason, &comma, syscall});
    } else {
        adamic_string space = ADAMIC_STRING(" ");
        message = adamic_string_concat(10, (adamic_string *const[]){prefix, &colon, reason, &comma, syscall, &space, &quote, (adamic_string *)from, &arrow, (adamic_string *)to});
        adamic_string *closed = adamic_string_concat(2, (adamic_string *const[]){message, &quote});
        adamic_release(message);
        message = closed;
    }
    adamic_thrown = adamic_object_new(&shape);
    adamic_thrown->slots[0].reference = &error_name;
    adamic_thrown->slots[1].reference = message;
    adamic_thrown->slots[2].reference = prefix;
    adamic_release(reason);
    adamic_release(syscall);
}

adamic_string *adamic_node_cwd(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_node_wasi_cwd();
#endif
    if (current_directory != NULL) { return adamic_retain(current_directory); }
    adamic_string *result = adamic_host_cwd();
    if (result == NULL) {
        host_directory_error(errno, "uv_cwd", NULL, NULL);
        return NULL;
    }
#ifdef ADAMIC_TARGET_WASI
    // Preview 1 keeps a path string, not the live directory handle.
    struct stat information;
    if (stat(".", &information) != 0) {
        adamic_release(result);
        adamic_panic("wasm32-wasi: process.cwd requires directory handles when saved cwd is unavailable", sizeof "wasm32-wasi: process.cwd requires directory handles when saved cwd is unavailable" - 1);
    }
#endif
    current_directory = result;
    return adamic_retain(result);
}

adamic_array *adamic_node_argv(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_panic("wasm32-wasi: process.argv requires executable paths", sizeof "wasm32-wasi: process.argv requires executable paths" - 1);
#endif
    if (arguments == NULL) {
        arguments = adamic_host_argv();
        if (arguments == NULL) {
            adamic_panic("process.argv: startup executable path unavailable", sizeof "process.argv: startup executable path unavailable" - 1);
        }
    }
    return adamic_retain(arguments);
}

adamic_array *adamic_node_exec_argv(void) {
    // Native has no VM flags; retain Node's array identity in the library adapter.
    if (execution_arguments == NULL) { execution_arguments = adamic_host_exec_argv(); }
    return adamic_retain(execution_arguments);
}

adamic_maybe_number adamic_node_columns(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_panic("wasm32-wasi: process.stdout.columns requires terminal size", sizeof "wasm32-wasi: process.stdout.columns requires terminal size" - 1);
#else
    adamic_maybe_number result = {0};
    if (isatty(STDOUT_FILENO)) {
        struct winsize size = {0};
        result.present = true;
        if (ioctl(STDOUT_FILENO, TIOCGWINSZ, &size) == 0) { result.number = size.ws_col; }
    }
    return result;
#endif
}

bool adamic_node_stdout_write(const adamic_string *text) {
    return adamic_write_raw(adamic_stdout, text);
}

adamic_object *adamic_node_memory_usage(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_panic("wasm32-wasi: process.memoryUsage requires allocator observations", sizeof "wasm32-wasi: process.memoryUsage requires allocator observations" - 1);
#endif
    adamic_host_memory observations;
    if (!adamic_host_memory_usage(&observations)) {
        adamic_panic("process.memoryUsage: host allocator observations unavailable", sizeof "process.memoryUsage: host allocator observations unavailable" - 1);
    }
    static const char *const memory_names[] = {"heapUsed"};
    static const bool references[] = {false};
    static const adamic_field_kind kinds[] = {adamic_field_number};
    static const adamic_shape shape = {1, memory_names, references, NULL, kinds};
    adamic_object *result = adamic_object_new(&shape);
    result->slots[0].number = observations.heapUsed;
    return result;
}

static adamic_object *entry(const adamic_string *name, bool measure, double start, double duration) {
    static const char *const entry_names[] = {"name", "entryType", "startTime", "duration"};
    static const bool references[] = {true, true, false, false};
    static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_number, adamic_field_number};
    static const adamic_shape shape = {4, entry_names, references, NULL, kinds};
    static adamic_string mark_type = ADAMIC_STRING("mark"), measure_type = ADAMIC_STRING("measure");
    adamic_object *result = adamic_object_new(&shape);
    result->slots[0].reference = adamic_retain((adamic_string *)name);
    result->slots[1].reference = measure ? &measure_type : &mark_type;
    result->slots[2].number = start;
    result->slots[3].number = duration;
    return result;
}

adamic_object *adamic_node_mark(const adamic_string *name) {
    double now = adamic_node_performance_now();
    mark_record *record = malloc(sizeof *record + name->length);
    if (record == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
    record->next = marks;
    record->length = name->length;
    record->time = now;
    memcpy(record->bytes, name->bytes, name->length);
    marks = record;
    return entry(name, false, now, 0);
}

static double mark_time(const adamic_string *name, double fallback) {
    if (name == NULL) { return fallback; }
    for (mark_record *record = marks; record != NULL; record = record->next) {
        if (record->length == name->length && memcmp(record->bytes, name->bytes, name->length) == 0) { return record->time; }
    }
    adamic_string prefix = ADAMIC_STRING("The \""), suffix = ADAMIC_STRING("\" performance mark has not been set");
    static adamic_string name_error = ADAMIC_STRING("SyntaxError");
    adamic_string *message = adamic_string_concat(3, (adamic_string *const[]){&prefix, (adamic_string *)name, &suffix});
    adamic_thrown = adamic_error_new(message);
    adamic_thrown->slots[0].reference = &name_error;
    adamic_release(message);
    return 0;
}

adamic_object *adamic_node_measure(const adamic_string *name, const adamic_string *start, const adamic_string *end) {
    // Node resolves the end mark first, including which missing-mark error wins.
    double finish = mark_time(end, adamic_node_performance_now());
    if (adamic_thrown != NULL) { return NULL; }
    double begin = mark_time(start, 0);
    if (adamic_thrown != NULL) { return NULL; }
    return entry(name, true, begin, finish - begin);
}

void adamic_node_clear_marks(const adamic_string *name) { clear_marks(name); }
void adamic_node_clear_measures(const adamic_string *name) { (void)name; }

#ifndef ADAMIC_TARGET_WASI
static adamic_value set_blocking(adamic_closure *self, adamic_value *values) {
    (void)self;
    adamic_output_flush();
    int flags = fcntl(STDOUT_FILENO, F_GETFL);
    if (flags >= 0) {
        if (values[0].boolean) { flags &= ~O_NONBLOCK; }
        else { flags |= O_NONBLOCK; }
        (void)fcntl(STDOUT_FILENO, F_SETFL, flags);
    }
    return (adamic_value){.reference = NULL};
}

#endif

adamic_object *adamic_node_stdout_handle(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_panic("wasm32-wasi: process.stdout._handle requires terminal flag control", sizeof "wasm32-wasi: process.stdout._handle requires terminal flag control" - 1);
#else
    struct stat info;
    if (fstat(STDOUT_FILENO, &info) != 0 || S_ISREG(info.st_mode)) { return NULL; }
    static const char *const handle_names[] = {"setBlocking"};
    static const bool references[] = {true};
    static const adamic_field_kind kinds[] = {adamic_field_reference};
    static const adamic_shape shape = {1, handle_names, references, NULL, kinds};
    static adamic_closure method = {.heap = {0, adamic_kind_closure, 0}, .code = set_blocking, .count = 0};
    adamic_object *handle = adamic_object_new(&shape);
    handle->slots[0].reference = &method;
    return handle;
#endif
}

// Node's environment uses NUL-terminated UTF-8, replacing lone UTF-16 surrogates.
static char *environment_text(const adamic_string *text) {
    char *bytes = malloc(text->length + 1);
    if (bytes == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
    memcpy(bytes, text->bytes, text->length);
    for (size_t at = 0; at + 3 <= text->length; at++) {
        if ((unsigned char)bytes[at] == 0xed && (unsigned char)bytes[at + 1] >= 0xa0) {
            memcpy(bytes + at, "\xef\xbf\xbd", 3);
            at += 2;
        }
    }
    bytes[text->length] = 0;
    return bytes;
}

void adamic_node_chdir(const adamic_string *directory) {
#ifdef ADAMIC_TARGET_WASI
    adamic_node_wasi_cwd();
#endif
    char *path = environment_text(directory);
#ifdef ADAMIC_TARGET_WASI
    // wasi-libc's cwd is virtual. Enter the physical path so later cwd() and
    // relative '..' preserve Node's kernel order after following a symlink.
    if (path[0] != '/') {
        struct stat information;
        if (stat(".", &information) != 0) {
            free(path);
            adamic_panic("wasm32-wasi: process.chdir requires directory handles when saved cwd is unavailable", sizeof "wasm32-wasi: process.chdir requires directory handles when saved cwd is unavailable" - 1);
        }
    }
    char *resolved = realpath(path, NULL);
    int changed = resolved == NULL ? -1 : chdir(resolved);
    int change_error = errno;
    free(resolved);
    errno = change_error;
#else
    int changed = chdir(path);
#endif
    if (changed == 0) {
        free(path);
        if (current_directory != NULL) { adamic_release(current_directory); current_directory = NULL; }
        return;
    }
    int error = errno;
    adamic_string *target = adamic_decode_utf8((const unsigned char *)path, strlen(path));
    free(path);
    adamic_string *from = adamic_node_cwd();
    if (adamic_thrown != NULL) { adamic_release(target); return; }
    host_directory_error(error, "chdir", from, target);
    adamic_release(target);
    adamic_release(from);
}

static adamic_value performance_now_method(adamic_closure *self, adamic_value *args) { (void)self; (void)args; return (adamic_value){.number = adamic_node_performance_now()}; }
static adamic_value performance_mark_method(adamic_closure *self, adamic_value *args) { (void)self; return (adamic_value){.reference = adamic_node_mark(args[0].reference)}; }
static adamic_value performance_measure_method(adamic_closure *self, adamic_value *args) { (void)self; return (adamic_value){.reference = adamic_node_measure(args[0].reference, args[1].reference, args[2].reference)}; }
static adamic_value performance_clear_marks_method(adamic_closure *self, adamic_value *args) { (void)self; adamic_node_clear_marks(args[0].reference); return (adamic_value){.reference = NULL}; }
static adamic_value performance_clear_measures_method(adamic_closure *self, adamic_value *args) { (void)self; adamic_node_clear_measures(args[0].reference); return (adamic_value){.reference = NULL}; }
static adamic_closure performance_now_closure = {.heap = {0, adamic_kind_closure, 0}, .code = performance_now_method, .count = 0};
static adamic_closure performance_mark_closure = {.heap = {0, adamic_kind_closure, 0}, .code = performance_mark_method, .count = 0};
static adamic_closure performance_measure_closure = {.heap = {0, adamic_kind_closure, 0}, .code = performance_measure_method, .count = 0};
static adamic_closure performance_clear_marks_closure = {.heap = {0, adamic_kind_closure, 0}, .code = performance_clear_marks_method, .count = 0};
static adamic_closure performance_clear_measures_closure = {.heap = {0, adamic_kind_closure, 0}, .code = performance_clear_measures_method, .count = 0};
static adamic_closure *const performance_methods[] = {&performance_now_closure, &performance_mark_closure, &performance_measure_closure, &performance_clear_marks_closure, &performance_clear_measures_closure};

adamic_object *adamic_node_performance(void) {
    if (performance_object == NULL) {
        static const char *const performance_names[] = {"timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures"};
        static const bool references[] = {false, true, true, true, true, true};
        static const adamic_field_kind kinds[] = {adamic_field_number, adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_reference};
        static const adamic_shape shape = {6, performance_names, references, NULL, kinds};
        performance_object = adamic_object_new(&shape);
        performance_object->slots[0].number = adamic_host_time_origin();
        for (size_t index = 0; index < 5; index++) { performance_object->slots[index+1].reference = performance_methods[index]; }
    }
    return adamic_retain(performance_object);
}

bool adamic_node_performance_method(const adamic_closure *closure) {
    if (closure == NULL) { return false; }
    for (size_t index = 0; index < 5; index++) { if (closure == performance_methods[index]) { return true; } }
    return false;
}

// Only the built-in closures need their omitted string arguments supplied. User closures keep
// their existing calling convention, and void views must release built-in entry return values.
adamic_value adamic_node_performance_invoke(adamic_closure *closure, adamic_value *args, size_t count, bool discard) {
    for (size_t index = 0; index < 5; index++) {
        if (closure == performance_methods[index]) {
            adamic_value padded[3] = {{.reference = NULL}, {.reference = NULL}, {.reference = NULL}};
            for (size_t at = 0; at < count && at < 3; at++) { padded[at] = args[at]; }
            adamic_value result = closure->code(closure, padded);
            if (discard) {
                // Main's discard arm may still inspect the returned word. Built-in void
                // calls hand it zero, after releasing the owned mark/measure result.
                if (index == 1 || index == 2) { adamic_release(result.reference); }
                return (adamic_value){.reference = NULL};
            }
            return result;
        }
    }
    return closure->code(closure, args);
}

// libuv/Node use the first nonempty POSIX temp variable and remove one trailing slash.
adamic_string *adamic_node_tmpdir(void) {
    adamic_string names[] = {ADAMIC_STRING("TMPDIR"), ADAMIC_STRING("TMP"), ADAMIC_STRING("TEMP")};
    adamic_string *value = NULL;
    for (size_t index = 0; index < 3; index++) {
        value = adamic_host_environment(&names[index]);
        if (value != NULL && value->length != 0) { break; }
        adamic_release(value);
        value = NULL;
    }
    if (value == NULL) { return adamic_decode_utf8((const unsigned char *)"/tmp", 4); }
    size_t length = value->length;
    if (length > 1 && value->bytes[length - 1] == '/') { length--; }
    if (length == value->length) { return value; }
    adamic_string *result = adamic_string_slice_bytes(value, 0, length);
    adamic_release(value);
    return result;
}
