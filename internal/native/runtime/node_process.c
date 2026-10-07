// node_process.c: synchronous compiler host observations and performance entries.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
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
    if (arguments != NULL) { adamic_release(arguments); }
    if (execution_arguments != NULL) { adamic_release(execution_arguments); }
    if (current_directory != NULL) { adamic_release(current_directory); }
    current_directory = NULL;
    arguments = NULL;
    execution_arguments = NULL;
}

void adamic_node_process_start(int count, char **values) {
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

double adamic_node_pid(void) { return (double)getpid(); }

adamic_string *adamic_node_platform(void) {
#if defined(__APPLE__)
    static adamic_string platform = ADAMIC_STRING("darwin");
#elif defined(__linux__)
    static adamic_string platform = ADAMIC_STRING("linux");
#else
#error Unsupported Node host platform
#endif
    return &platform;
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
        // Catchable system errors need a code-bearing Error shape and exceptional CFG edges.
        // Lowering refuses a try around cwd until that shared integration exists.
        const char *message = error == ENOENT ? "Error: ENOENT: no such file or directory, uv_cwd" : "Error: cannot read current directory";
        adamic_panic(message, strlen(message));
    }
}

static adamic_string *executable_path(void) {
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
    adamic_maybe_number result = {0};
    if (isatty(STDOUT_FILENO)) {
        struct winsize size = {0};
        result.present = true;
        if (ioctl(STDOUT_FILENO, TIOCGWINSZ, &size) == 0) { result.number = size.ws_col; }
    }
    return result;
}

bool adamic_node_stdout_write(const adamic_string *text) {
    return adamic_write_raw(adamic_stdout, text);
}

adamic_object *adamic_node_memory_usage(void) {
    static const char *const names[] = {"heapUsed"};
    static const bool references[] = {false};
    static const adamic_shape shape = {1, names, references, NULL};
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
    static const char *const names[] = {"name", "entryType", "startTime", "duration"};
    static const bool references[] = {true, true, false, false};
    static const adamic_shape shape = {4, names, references, NULL};
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
    adamic_string prefix = ADAMIC_STRING("SyntaxError: The \"");
    adamic_string suffix = ADAMIC_STRING("\" performance mark has not been set");
    adamic_string *message = adamic_string_concat(3, (adamic_string *const[]){&prefix, (adamic_string *)name, &suffix});
    adamic_panic(message->bytes, message->length);
}

adamic_object *adamic_node_measure(const adamic_string *name, const adamic_string *start, const adamic_string *end) {
    // Node resolves the end mark first, including which missing-mark error wins.
    double finish = mark_time(end, adamic_node_performance_now());
    double begin = mark_time(start, 0);
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
    static const char *const names[] = {"setBlocking"};
    static const bool references[] = {true};
    static const adamic_shape shape = {1, names, references, NULL};
    static adamic_closure method = {{0, adamic_kind_closure, 0}, set_blocking, 0};
    adamic_object *handle = adamic_object_new(&shape);
    handle->slots[0].reference = &method;
    return handle;
}
