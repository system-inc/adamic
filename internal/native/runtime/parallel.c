// parallel.c: structured fork-join, ranges in worker deques, and helping at every join.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "parallel.h"
#include <errno.h>
#include <pthread.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define GRAIN 256

typedef struct scope {
	adamic_array *items;
	adamic_closure *work;
	adamic_array *results;
	size_t completed;
	size_t exception_index;
	adamic_object *exception;
} scope;

typedef struct range {
	struct range *previous;
	struct range *next;
	scope *scope;
	size_t from;
	size_t end;
} range;

typedef struct worker {
	pthread_t thread;
	range *first;
	range *last;
} worker;

static pthread_once_t started = PTHREAD_ONCE_INIT;
static pthread_mutex_t scheduler = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t changed = PTHREAD_COND_INITIALIZER;
static worker *workers;
static size_t thread_count;
static size_t created;
static bool stopping;
static _Thread_local size_t worker_index;

// On Linux use the process's cgroup v2 quota when it reduces the online CPU count. On systems
// without that file, including macOS, online CPUs decide. An explicit override always wins.
static size_t available_threads(void) {
	long online = sysconf(_SC_NPROCESSORS_ONLN);
	size_t count = online > 0 ? (size_t)online : 1;
#ifdef __linux__
	FILE *quota_file = fopen("/sys/fs/cgroup/cpu.max", "r");
	if (quota_file != NULL) {
		unsigned long long quota, period;
		if (fscanf(quota_file, "%llu %llu", &quota, &period) == 2 && quota > 0 && period > 0) {
			unsigned long long limited = quota / period + (quota % period != 0);
			if (limited < count) { count = (size_t)limited; }
		}
		fclose(quota_file);
	}
#endif
	const char *override = getenv("ADAMIC_THREADS");
	if (override != NULL) {
		char *end;
		errno = 0;
		unsigned long requested = strtoul(override, &end, 10);
		if (errno != 0 || *override == '\0' || *override == '-' || *end != '\0' || requested == 0 || requested > 4096) {
			adamic_panic("ADAMIC_THREADS must be an integer from 1 to 4096", sizeof "ADAMIC_THREADS must be an integer from 1 to 4096" - 1);
		}
		count = (size_t)requested;
	}
	return count;
}

static range *new_range(scope *scope, size_t from, size_t end) {
	range *task = malloc(sizeof *task);
	if (task == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
	*task = (range){.scope = scope, .from = from, .end = end};
	return task;
}

// Deque links and all scope completion/exception state are owned by the scheduler mutex.
static void append(worker *worker, range *task) {
	task->previous = worker->last;
	task->next = NULL;
	if (worker->last != NULL) { worker->last->next = task; } else { worker->first = task; }
	worker->last = task;
}

static void remove_range(worker *worker, range *task) {
	if (task->previous != NULL) { task->previous->next = task->next; } else { worker->first = task->next; }
	if (task->next != NULL) { task->next->previous = task->previous; } else { worker->last = task->previous; }
}

// Claim a chunk from our newest range. A thief takes the oldest range and splits it in half,
// leaving its victim work. The remaining range stays stealable even while its owner calls work.
static bool claim(scope **scope_out, size_t *from_out, size_t *end_out) {
	worker *own = &workers[worker_index];
	range *task = own->last;
	if (task == NULL) {
		for (size_t distance = 1; distance < thread_count; distance++) {
			worker *victim = &workers[(worker_index + distance) % thread_count];
			range *oldest = victim->first;
			if (oldest == NULL) { continue; }
			size_t length = oldest->end - oldest->from;
			if (length > GRAIN) {
				size_t middle = oldest->from + length / 2;
				task = new_range(oldest->scope, middle, oldest->end);
				oldest->end = middle;
			} else {
				task = oldest;
				remove_range(victim, oldest);
			}
			append(own, task);
			break;
		}
	}
	if (task == NULL) { return false; }
	*scope_out = task->scope;
	*from_out = task->from;
	*end_out = task->end - task->from > GRAIN ? task->from + GRAIN : task->end;
	task->from = *end_out;
	if (task->from == task->end) { remove_range(own, task); free(task); }
	return true;
}

static void execute_range(scope *scope, size_t from, size_t end) {
	adamic_object *outer_exception = adamic_thrown;
	adamic_thrown = NULL;
	for (size_t index = from; index < end; index++) {
		adamic_value arguments[] = {scope->items->elements[index], {.number = (double)index}};
		adamic_value result = scope->work->code(scope->work, arguments);
		if (adamic_thrown != NULL) {
			adamic_object *error = adamic_thrown;
			adamic_thrown = NULL;
			adamic_share(error);
			pthread_mutex_lock(&scheduler);
			if (index < scope->exception_index) {
				adamic_object *previous = scope->exception;
				scope->exception = error;
				scope->exception_index = index;
				error = previous;
			}
			pthread_mutex_unlock(&scheduler);
			adamic_release(error);
		} else {
			if (scope->results->references) { adamic_share(result.reference); }
			scope->results->elements[index] = result;
		}
	}
	adamic_thrown = outer_exception;
	pthread_mutex_lock(&scheduler);
	scope->completed += end - from;
	pthread_cond_broadcast(&changed);
	pthread_mutex_unlock(&scheduler);
}

static void *worker_main(void *given) {
	worker_index = (size_t)given;
	adamic_stack_thread_start();
	pthread_mutex_lock(&scheduler);
	for (;;) {
		scope *scope;
		size_t from, end;
		while (!stopping && !claim(&scope, &from, &end)) { pthread_cond_wait(&changed, &scheduler); }
		if (stopping) { break; }
		pthread_mutex_unlock(&scheduler);
		execute_range(scope, from, end);
		pthread_mutex_lock(&scheduler);
	}
	pthread_mutex_unlock(&scheduler);
	adamic_heap_thread_end();
	return NULL;
}

void adamic_parallel_shutdown(void) {
	if (created == 0) { return; }
	pthread_mutex_lock(&scheduler);
	stopping = true;
	pthread_cond_broadcast(&changed);
	pthread_mutex_unlock(&scheduler);
	for (size_t index = 1; index <= created; index++) { pthread_join(workers[index].thread, NULL); }
	free(workers);
	workers = NULL;
	created = 0;
}

static void start(void) {
	thread_count = available_threads();
	// This path creates no threads, not even an idle worker.
	if (thread_count == 1) { return; }
	workers = calloc(thread_count, sizeof *workers);
	if (workers == NULL) { adamic_panic("out of memory", sizeof "out of memory" - 1); }
	pthread_attr_t attributes;
	pthread_attr_init(&attributes);
	if (pthread_attr_setstacksize(&attributes, (size_t)8 << 20) != 0) { adamic_panic("cannot set worker stack size", sizeof "cannot set worker stack size" - 1); }
	// Only the caller handles termination signals; its stdout state is quiescent at every join.
	sigset_t signals, saved;
	sigemptyset(&signals); sigaddset(&signals, SIGTERM); sigaddset(&signals, SIGINT); sigaddset(&signals, SIGHUP);
	pthread_sigmask(SIG_BLOCK, &signals, &saved);
	for (size_t index = 1; index < thread_count; index++) {
		if (pthread_create(&workers[index].thread, &attributes, worker_main, (void *)index) != 0) {
			adamic_panic("cannot start parallel worker", sizeof "cannot start parallel worker" - 1);
		}
		created++;
	}
	pthread_sigmask(SIG_SETMASK, &saved, NULL);
	pthread_attr_destroy(&attributes);
	atexit(adamic_parallel_shutdown);
}

size_t adamic_parallel_threads(void) { pthread_once(&started, start); return thread_count; }
size_t adamic_parallel_workers(void) { pthread_once(&started, start); return created; }

adamic_array *adamic_parallel_map(adamic_array *items, adamic_closure *work) {
	pthread_once(&started, start);
	adamic_share(items);
	adamic_share(work);
	adamic_array *results = adamic_array_new(items->length, work->result_references);
	results->length = items->length;
	if (items->length != 0) { memset(results->elements, 0, items->length * sizeof *results->elements); }
	if (thread_count == 1) {
		for (size_t index = 0; index < items->length; index++) {
			adamic_value arguments[] = {items->elements[index], {.number = (double)index}};
			adamic_value result = work->code(work, arguments);
			if (adamic_thrown != NULL) { adamic_release(results); return NULL; }
			if (results->references) { adamic_share(result.reference); }
			results->elements[index] = result;
		}
		return results;
	}
	scope state = {.items = items, .work = work, .results = results, .exception_index = SIZE_MAX};
	pthread_mutex_lock(&scheduler);
	if (items->length != 0) {
		append(&workers[worker_index], new_range(&state, 0, items->length));
		pthread_cond_broadcast(&changed);
	}
	// Nested callers help run their own children or stolen work; no worker blocks with runnable work.
	while (state.completed != items->length) {
		scope *next;
		size_t from, end;
		if (!claim(&next, &from, &end)) { pthread_cond_wait(&changed, &scheduler); continue; }
		pthread_mutex_unlock(&scheduler);
		execute_range(next, from, end);
		pthread_mutex_lock(&scheduler);
	}
	pthread_mutex_unlock(&scheduler);
	if (state.exception != NULL) {
		adamic_release(results);
		adamic_thrown = state.exception;
		return NULL;
	}
	return results;
}
