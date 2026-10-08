// node_process.c: synchronous compiler host observations and performance entries.
#define _POSIX_C_SOURCE 200809L
// macOS hides malloc/malloc.h's zone types once _POSIX_C_SOURCE is set, unless Darwin's
// own extensions are asked for too. glibc ignores this macro.
#define _DARWIN_C_SOURCE
#include "adamic.h"
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
#if defined(__linux__)
#include <malloc.h>
#elif defined(__APPLE__)
#include <malloc/malloc.h>
#include <mach-o/dyld.h>
#endif

#if defined(__has_feature)
#if __has_feature(address_sanitizer)
#define ADAMIC_HOST_SANITIZED 1
#include <sanitizer/allocator_interface.h>
#endif
#endif
#ifndef ADAMIC_HOST_SANITIZED
#define ADAMIC_HOST_SANITIZED 0
#endif

static int saved_count;
static char **saved_values;
static double monotonic_origin, epoch_origin;
static adamic_array *arguments, *execution_arguments;
static adamic_string *current_directory;
static adamic_object *performance_object;

static double milliseconds(clockid_t clock) {
    struct timespec value;
    if (clock_gettime(clock, &value) != 0) {
        static const char message[] = "cannot read host clock";
        adamic_panic(message, sizeof message - 1);
    }
    return (double)value.tv_sec * 1000 + (double)value.tv_nsec / 1000000;
}

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

void adamic_node_process_start(int count, char **values) {
#ifdef ADAMIC_TARGET_WASI
    // Preview 1 has no initial cwd. A command host can explicitly supply one
    // inside its preopens; otherwise wasi-libc's virtual cwd remains '/'.
    const char *directory = getenv("ADAMIC_WASI_CWD");
    if (directory != NULL && chdir(directory) != 0) {
        adamic_panic("wasm32-wasi: cannot enter supplied working directory", sizeof "wasm32-wasi: cannot enter supplied working directory" - 1);
    }
#endif
    saved_count = count;
    saved_values = values;
    monotonic_origin = milliseconds(CLOCK_MONOTONIC);
    epoch_origin = milliseconds(CLOCK_REALTIME);
    atexit(finish_host);
}

double adamic_node_performance_now(void) { return milliseconds(CLOCK_MONOTONIC) - monotonic_origin; }
double adamic_node_time_origin(void) { return epoch_origin; }
adamic_string *adamic_node_eol(void) {
    static adamic_string newline = ADAMIC_STRING("\n");
    return &newline;
}

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
    adamic_thrown = adamic_builtin_error_new(0, message);
    adamic_thrown->slots[2].reference = prefix;
    adamic_release(message);
    adamic_release(reason);
    adamic_release(syscall);
}

adamic_string *adamic_node_cwd(void) {
    if (current_directory != NULL) { return adamic_retain(current_directory); }
    size_t capacity = 256;
    for (;;) {
        char *buffer = malloc(capacity);
        if (buffer == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
        if (getcwd(buffer, capacity) != NULL) {
            adamic_string *result = adamic_decode_utf8((const unsigned char *)buffer, strlen(buffer));
            free(buffer);
            current_directory = result;
            return adamic_retain(result);
        }
        int error = errno;
        free(buffer);
        if (error == ERANGE) { capacity *= 2; continue; }
        host_directory_error(error, "uv_cwd", NULL, NULL);
        return NULL;
    }
}

static adamic_string *executable_path(void) {
#ifdef ADAMIC_TARGET_WASI
    adamic_panic("wasm32-wasi: process.argv requires executable paths", sizeof "wasm32-wasi: process.argv requires executable paths" - 1);
#endif
#if defined(__linux__)
    size_t capacity = 256;
    for (;;) {
        char *buffer = malloc(capacity);
        if (buffer == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
        ssize_t length = readlink("/proc/self/exe", buffer, capacity);
        if (length >= 0 && (size_t)length < capacity) {
            adamic_string *result = adamic_decode_utf8((const unsigned char *)buffer, (size_t)length);
            free(buffer);
            return result;
        }
        free(buffer);
        if (length < 0) { break; }
        capacity *= 2;
    }
#elif defined(__APPLE__)
    uint32_t capacity = 0;
    (void)_NSGetExecutablePath(NULL, &capacity);
    char *buffer = malloc(capacity);
    if (buffer != NULL && _NSGetExecutablePath(buffer, &capacity) == 0) {
        adamic_string *result = adamic_decode_utf8((const unsigned char *)buffer, strlen(buffer));
        free(buffer);
        return result;
    }
    free(buffer);
#endif
    const char *fallback = saved_count > 0 ? saved_values[0] : "";
    return adamic_decode_utf8((const unsigned char *)fallback, strlen(fallback));
}

adamic_array *adamic_node_argv(void) {
    if (arguments == NULL) {
        size_t count = saved_count > 0 ? (size_t)saved_count + 1 : 2;
        arguments = adamic_array_new(count, true);
        adamic_string *path = executable_path();
        adamic_array_push(arguments, (adamic_value){.reference = path});
        adamic_array_push(arguments, (adamic_value){.reference = adamic_retain(path)});
        for (int index = 1; index < saved_count; index++) {
            const char *value = saved_values[index];
            adamic_array_push(arguments, (adamic_value){.reference = adamic_decode_utf8((const unsigned char *)value, strlen(value))});
        }
    }
    return adamic_retain(arguments);
}

adamic_array *adamic_node_exec_argv(void) {
    // A native executable has no Node VM flags. User arguments belong to argv, even --prof.
    if (execution_arguments == NULL) { execution_arguments = adamic_array_new(0, true); }
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
    static const char *const memory_names[] = {"heapUsed"};
    static const bool references[] = {false};
    static const adamic_shape shape = {1, memory_names, references, NULL};
    adamic_object *result = adamic_object_new(&shape);
    double used = 0;
#if ADAMIC_HOST_SANITIZED
    used = (double)__sanitizer_get_current_allocated_bytes();
#elif defined(__linux__)
    struct mallinfo2 info = mallinfo2();
    used = (double)info.uordblks + (double)info.hblkhd;
#elif defined(__APPLE__)
    malloc_statistics_t info;
    malloc_zone_statistics(NULL, &info);
    used = (double)info.size_in_use;
#endif
    result->slots[0].number = used;
    return result;
}

static adamic_object *entry(const adamic_string *name, bool measure, double start, double duration) {
    static const char *const entry_names[] = {"name", "entryType", "startTime", "duration"};
    static const bool references[] = {true, true, false, false};
    static const adamic_shape shape = {4, entry_names, references, NULL};
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
    adamic_string *message = adamic_string_concat(3, (adamic_string *const[]){&prefix, (adamic_string *)name, &suffix});
    // A missing performance mark is a DOMException named SyntaxError on Node.
    adamic_thrown = adamic_builtin_error_new(8, message);
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

adamic_object *adamic_node_stdout_handle(void) {
    struct stat info;
    if (fstat(STDOUT_FILENO, &info) != 0 || S_ISREG(info.st_mode)) { return NULL; }
    static const char *const handle_names[] = {"setBlocking"};
    static const bool references[] = {true};
    static const adamic_shape shape = {1, handle_names, references, NULL};
    static adamic_closure method = {{0, adamic_kind_closure, 0}, set_blocking, 0};
    adamic_object *handle = adamic_object_new(&shape);
    handle->slots[0].reference = &method;
    return handle;
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

void adamic_node_environment_set(const adamic_string *name, const adamic_string *value) {
    char *key = environment_text(name), *text = environment_text(value);
    // Node ignores invalid names (empty or containing '=') rather than throwing.
    (void)setenv(key, text, 1);
    free(text);
    free(key);
}

bool adamic_node_environment_delete(const adamic_string *name) {
    char *key = environment_text(name);
    (void)unsetenv(key);
    free(key);
    return true;
}

void adamic_node_chdir(const adamic_string *directory) {
    char *path = environment_text(directory);
    if (chdir(path) == 0) {
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
static adamic_closure performance_now_closure = {{0, adamic_kind_closure, 0}, performance_now_method, 0};
static adamic_closure performance_mark_closure = {{0, adamic_kind_closure, 0}, performance_mark_method, 0};
static adamic_closure performance_measure_closure = {{0, adamic_kind_closure, 0}, performance_measure_method, 0};
static adamic_closure performance_clear_marks_closure = {{0, adamic_kind_closure, 0}, performance_clear_marks_method, 0};
static adamic_closure performance_clear_measures_closure = {{0, adamic_kind_closure, 0}, performance_clear_measures_method, 0};
static adamic_closure *const performance_methods[] = {&performance_now_closure, &performance_mark_closure, &performance_measure_closure, &performance_clear_marks_closure, &performance_clear_measures_closure};

adamic_object *adamic_node_performance(void) {
    if (performance_object == NULL) {
        static const char *const performance_names[] = {"timeOrigin", "now", "mark", "measure", "clearMarks", "clearMeasures"};
        static const bool references[] = {false, true, true, true, true, true};
        static const adamic_shape shape = {6, performance_names, references, NULL};
        performance_object = adamic_object_new(&shape);
        performance_object->slots[0].number = epoch_origin;
        for (size_t index = 0; index < 5; index++) { performance_object->slots[index+1].reference = performance_methods[index]; }
    }
    return adamic_retain(performance_object);
}

// Only the built-in closures need their omitted string arguments supplied. User closures keep
// their existing calling convention, and void views must release built-in entry return values.
adamic_value adamic_node_performance_invoke(adamic_closure *closure, adamic_value *args, size_t count, bool discard) {
    for (size_t index = 0; index < 5; index++) {
        if (closure == performance_methods[index]) {
            adamic_value padded[3] = {{.reference = NULL}, {.reference = NULL}, {.reference = NULL}};
            for (size_t at = 0; at < count && at < 3; at++) { padded[at] = args[at]; }
            adamic_value result = closure->code(closure, padded);
            if (discard && (index == 1 || index == 2)) { adamic_release(result.reference); result.reference = NULL; }
            return result;
        }
    }
    return closure->code(closure, args);
}

// libuv/Node use the first nonempty POSIX temp variable and remove one trailing slash.
adamic_string *adamic_node_tmpdir(void) {
    const char *value = getenv("TMPDIR");
    if (value == NULL || value[0] == 0) { value = getenv("TMP"); }
    if (value == NULL || value[0] == 0) { value = getenv("TEMP"); }
    if (value == NULL || value[0] == 0) { value = "/tmp"; }
    size_t length = strlen(value);
    if (length > 1 && value[length - 1] == '/') { length--; }
    return adamic_decode_utf8((const unsigned char *)value, length);
}

void adamic_node_error(adamic_string *name, adamic_string *message, adamic_string *code) {
    static adamic_string names[] = {
        ADAMIC_STRING("Error"), ADAMIC_STRING("TypeError"), ADAMIC_STRING("SyntaxError"),
        ADAMIC_STRING("RangeError"), ADAMIC_STRING("ReferenceError"), ADAMIC_STRING("EvalError"),
        ADAMIC_STRING("URIError"), ADAMIC_STRING("SystemError")
    };
    int kind = 0;
    for (int candidate = 1; candidate < 8; candidate++) {
        if (adamic_string_equal(name, &names[candidate])) { kind = candidate; break; }
    }
    adamic_thrown = adamic_builtin_error_new(kind, message);
    adamic_release(adamic_thrown->slots[0].reference);
    adamic_thrown->slots[0].reference = adamic_retain(name);
    adamic_thrown->slots[2].reference = adamic_retain(code);
    adamic_error_tag(adamic_thrown);
}
