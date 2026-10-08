// host_runtime.c: process observations and clocks behind library's checked Node imports.
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
#include "host_runtime.h"
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#if defined(__linux__)
#include <malloc.h>
#if defined(__GLIBC__)
#if __GLIBC_PREREQ(2, 33)
#define HOST_MALLINFO2 1
#endif
#endif
#elif defined(__APPLE__)
#include <malloc/malloc.h>
#include <mach/mach.h>
#include <mach-o/dyld.h>
#endif
#if defined(__has_feature)
#if __has_feature(address_sanitizer)
#define HOST_ASAN 1
#include <sanitizer/allocator_interface.h>
#endif
#endif
#ifndef HOST_ASAN
#define HOST_ASAN 0
#endif

static int saved_count;
static char **saved_values;
static char *identity_bytes;
static size_t identity_length;
static int identity_error = EINVAL;
static bool identity_saved;
static bool capture_executable_identity(void);
static void release_identity(void) { free(identity_bytes); identity_bytes = NULL; }
static bool exit_present;
static int32_t exit_value;
static struct timespec monotonic_origin;
static double epoch_origin;
static _Atomic size_t backing_bytes;

static struct timespec clock_value(clockid_t clock) {
	struct timespec value;
	if (clock_gettime(clock, &value) != 0) {
		adamic_panic("cannot read host clock", sizeof "cannot read host clock" - 1);
	}
	return value;
}

__attribute__((constructor(101))) static void capture_clock_origin(void) {
	// Pair epoch with the midpoint of monotonic samples; subtract before converting to double.
	struct timespec before = clock_value(CLOCK_MONOTONIC);
	struct timespec epoch = clock_value(CLOCK_REALTIME);
	struct timespec after = clock_value(CLOCK_MONOTONIC);
	int64_t gap = ((int64_t)after.tv_sec - before.tv_sec) * INT64_C(1000000000) + after.tv_nsec - before.tv_nsec;
	monotonic_origin = before;
	monotonic_origin.tv_nsec += (long)(gap / 2);
	if (monotonic_origin.tv_nsec >= 1000000000) { monotonic_origin.tv_sec++; monotonic_origin.tv_nsec -= 1000000000; }
	epoch_origin = (double)epoch.tv_sec * 1000.0 + (double)epoch.tv_nsec / 1000000.0;
}

void adamic_host_start(int count, char **values) {
	saved_count = count; saved_values = values;
	if (!identity_saved) {
		identity_saved = true;
		if (!capture_executable_identity()) { identity_error = errno; }
		if (identity_bytes != NULL) { atexit(release_identity); }
	}
}

void adamic_host_set_exit_code(bool present, int32_t code) { exit_present = present; exit_value = code; }
bool adamic_host_exit_code(int32_t *code) { *code = exit_value; return exit_present; }
int adamic_host_exit_status(void) { return exit_present ? (int)exit_value : 0; }

double adamic_host_performance_now(void) {
	struct timespec now = clock_value(CLOCK_MONOTONIC);
	return ((double)now.tv_sec - (double)monotonic_origin.tv_sec) * 1000.0 +
		(double)(now.tv_nsec - monotonic_origin.tv_nsec) / 1000000.0;
}
double adamic_host_time_origin(void) { return epoch_origin; }
double adamic_host_date_now(void) {
	struct timespec now = clock_value(CLOCK_REALTIME);
	return (double)now.tv_sec * 1000.0 + (double)(now.tv_nsec / 1000000);
}

adamic_string *adamic_host_environment(const adamic_string *name) {
	// Node's C environment lookup truncates a name at NUL; encode lone surrogates as U+FFFD.
	char *bytes = malloc(name->length + 1);
	if (bytes == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
	memcpy(bytes, name->bytes, name->length);
	for (size_t at = 0; at + 2 < name->length; at++) {
		if ((unsigned char)bytes[at] == 0xed && (unsigned char)bytes[at + 1] >= 0xa0) {
			memcpy(bytes + at, "\xef\xbf\xbd", 3); at += 2;
		}
	}
	bytes[name->length] = 0;
	const char *value = getenv(bytes);
	free(bytes);
	return value == NULL ? NULL : adamic_decode_utf8((const unsigned char *)value, strlen(value));
}

adamic_string *adamic_host_cwd(void) {
	size_t capacity = 256;
	for (;;) {
		char *buffer = malloc(capacity);
		if (buffer == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
		if (getcwd(buffer, capacity) != NULL) {
			adamic_string *result = adamic_decode_utf8((const unsigned char *)buffer, strlen(buffer));
			free(buffer); return result;
		}
		int error = errno; free(buffer);
		if (error != ERANGE) { errno = error; return NULL; }
		capacity *= 2;
	}
}

static bool capture_executable_identity(void) {
#if defined(__linux__)
	size_t capacity = 256;
	for (;;) {
		char *buffer = malloc(capacity);
		if (buffer == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
		ssize_t length = readlink("/proc/self/exe", buffer, capacity);
		if (length >= 0 && (size_t)length < capacity) {
			identity_bytes = buffer; identity_length = (size_t)length; return true;
		}
		int error = errno; free(buffer);
		if (length < 0) { errno = error; return false; }
		capacity *= 2;
	}
#elif defined(__APPLE__)
	uint32_t capacity = 0;
	(void)_NSGetExecutablePath(NULL, &capacity);
	char *buffer = malloc(capacity);
	if (buffer == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
	if (_NSGetExecutablePath(buffer, &capacity) != 0) { free(buffer); errno = EIO; return false; }
	char *resolved = realpath(buffer, NULL);
	free(buffer);
	if (resolved == NULL) { return false; }
	identity_bytes = resolved; identity_length = strlen(resolved); return true;
#else
	errno = ENOSYS; return false;
#endif
}

adamic_string *adamic_host_exec_path(void) {
	if (identity_bytes == NULL) { errno = identity_error; return NULL; }
	return adamic_decode_utf8((const unsigned char *)identity_bytes, identity_length);
}

adamic_string *adamic_host_executing_file_path(void) { return adamic_host_exec_path(); }
adamic_array *adamic_host_argv(void) {
	adamic_string *path = adamic_host_exec_path();
	if (path == NULL) { return NULL; }
	adamic_array *array = adamic_array_new(saved_count > 0 ? (size_t)saved_count + 1 : 2, true);
	adamic_array_push(array, (adamic_value){.reference = path});
	adamic_array_push(array, (adamic_value){.reference = adamic_retain(path)});
	for (int index = 1; index < saved_count; index++) {
		adamic_array_push(array, (adamic_value){.reference = adamic_decode_utf8((const unsigned char *)saved_values[index], strlen(saved_values[index]))});
	}
	return array;
}
adamic_array *adamic_host_exec_argv(void) { return adamic_array_new(0, true); }
adamic_string *adamic_host_eol(void) { static adamic_string eol = ADAMIC_STRING("\n"); return &eol; }

void adamic_host_backing_add(size_t bytes) { atomic_fetch_add_explicit(&backing_bytes, bytes, memory_order_relaxed); }
void adamic_host_backing_drop(size_t bytes) { atomic_fetch_sub_explicit(&backing_bytes, bytes, memory_order_relaxed); }
bool adamic_host_memory_usage(adamic_host_memory *result) {
	*result = (adamic_host_memory){0};
#if defined(__linux__)
	FILE *stat = fopen("/proc/self/statm", "r");
	if (stat == NULL) { return false; }
	unsigned long total, resident;
	int fields = fscanf(stat, "%lu %lu", &total, &resident);
	fclose(stat);
	long page = sysconf(_SC_PAGESIZE);
	if (fields != 2 || page <= 0) { errno = EIO; return false; }
	result->rss = (double)resident * (double)page;
#elif defined(__APPLE__)
	mach_task_basic_info_data_t info;
	mach_msg_type_number_t count = MACH_TASK_BASIC_INFO_COUNT;
	if (task_info(mach_task_self(), MACH_TASK_BASIC_INFO, (task_info_t)&info, &count) != KERN_SUCCESS) { errno = EIO; return false; }
	result->rss = (double)info.resident_size;
#else
	errno = ENOSYS; return false;
#endif
#if HOST_ASAN
	result->heapTotal = (double)__sanitizer_get_heap_size();
	result->heapUsed = (double)__sanitizer_get_current_allocated_bytes();
#elif defined(HOST_MALLINFO2)
	struct mallinfo2 info = mallinfo2();
	result->heapTotal = (double)info.arena + (double)info.hblkhd;
	result->heapUsed = (double)info.uordblks + (double)info.hblkhd;
#elif defined(__APPLE__)
	malloc_statistics_t info;
	malloc_zone_statistics(NULL, &info);
	result->heapTotal = (double)info.size_allocated;
	result->heapUsed = (double)info.size_in_use;
#else
	errno = ENOSYS; return false;
#endif
	result->external = result->arrayBuffers = (double)atomic_load_explicit(&backing_bytes, memory_order_relaxed);
	return true;
}
